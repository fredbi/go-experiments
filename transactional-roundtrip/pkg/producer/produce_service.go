package producer

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"time"

	"github.com/fredbi/go-experiments/transactional-roundtrip/pkg/injected"
	natsembedded "github.com/fredbi/go-experiments/transactional-roundtrip/pkg/nats"
	"github.com/fredbi/go-experiments/transactional-roundtrip/pkg/repos"

	"github.com/fredbi/go-trace/log"
	"github.com/fredbi/go-trace/tracer"
	nats "github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

var (
	ErrProduceDB  = errors.New("could not write in DB")
	ErrProduceMsg = errors.New("message taken into account, but service temporily unavailable")
)

type Producer struct {
	ID                string
	rt                injected.Runtime
	publishedSubject  func(string) string
	subscribedSubject string
	nc                *nats.Conn // that's the connection to the NATS embedded server. TODO: perf - could use direct in-process connection

	settings
}

func New(rt injected.Runtime, id string) *Producer {
	return &Producer{
		ID: id,
		rt: rt,
	}
}

func (p Producer) Logger() log.Factory {
	return p.rt.Logger()
}

func (p Producer) createAndSendMessage(parentCtx context.Context, msg repos.Message) error {
	spanCtx, span, lg := tracer.StartSpan(parentCtx, p)
	defer span.End()

	ctx, cancel := context.WithTimeout(spanCtx, p.Producer.MsgProcessTimeout)
	defer cancel()

	lg = lg.With(zap.String("id", msg.ID))

	// encode the message as []byte, using gob
	payload, err := msg.Bytes()
	if err != nil {
		lg.Warn("could not marshal message",
			zap.String("outcome", "user input is discarded"),
			zap.Error(err),
		)

		return err
	}

	// write and commit to DB. After this point, an eventual delivery is guaranteed (possibly with replays)
	if err := p.rt.Repos().Messages().Create(ctx, msg); err != nil {
		if errors.Is(err, repos.ErrAlreadyProcessed) {
			lg.Warn("duplicate message entry detected",
				zap.String("outcome", "no new message is created, user input is discarded"),
				zap.Error(err),
			)
		} else {
			lg.Warn("could not write in DB",
				zap.String("outcome", "message not accepted in the system"),
				zap.Error(err),
			)
		}

		return errors.Join(
			ErrProduceDB,
			err,
		)
	}

	lg.Debug("message created by producer")

	// send message with subject "postings.{ConsumerID}" and reply-to
	post := nats.NewMsg(p.publishedSubject(msg.ConsumerID))
	post.Data = payload
	post.Reply = p.subscribedSubject
	post.Header.Add("trace_id", span.SpanContext().TraceID.String())
	post.Header.Add("span_id", span.SpanContext().SpanID.String())

	// In this mode, NATS operates as fire-and-forget: there is no ACK.
	// Errors come only from surface checks on the message.
	if err := p.nc.PublishMsg(post); err != nil {
		lg.Warn("could not publish",
			zap.String("outcome", "will be resubmitted later"),
			zap.Error(err),
		)
	}

	lg.Debug("message posted to consumer", zap.String("consumer_id", msg.ConsumerID), zap.Any("raw_msg", post))

	return nil
}

// subscriptionHandler processes responses from consumers.
//
// It updates the processing status then sends back a confirmation.
//
// Until confirmations are acknowledged by the consumer, it will keep redelivering responses.
func (p Producer) subscriptionHandler(incoming *nats.Msg) {
	spanCtx := natsembedded.SpanContextFromHeaders(context.Background(), incoming)
	parentCtx, cancel := context.WithTimeout(spanCtx, p.Producer.MsgProcessTimeout)
	defer cancel()

	ctx, span, lg := tracer.StartSpan(parentCtx, p)
	defer span.End()

	lg = lg.With(
		zap.String("subject", incoming.Subject),
		zap.Int("size", incoming.Size()),
	)
	defer span.End()

	buffer := bytes.NewReader(incoming.Data)
	var msg repos.Message
	dec := gob.NewDecoder(buffer)
	if err := dec.Decode(&msg); err != nil {
		lg.Warn("invalid message content",
			zap.String("outcome", "message thrown away"),
			zap.Error(err),
		)

		return
	}
	lg = lg.With(zap.String("id", msg.ID))

	// sanity check on message status protocol.
	// We expect only responses with a decided outcome.
	if msg.MessageStatus != repos.MessageStatusPosted {
		lg.Error("unexpected response status",
			zap.String("outcome", "message thrown away"),
			zap.Stringer("message_status", msg.MessageStatus),
			zap.Stringer("processing_status", msg.ProcessingStatus),
		)

		return
	}

	// sanity check on processing status: expect an outcome to have been decided
	switch msg.ProcessingStatus {
	case repos.ProcessingStatusOK:
		lg.Info("message has been processed OK")
	case repos.ProcessingStatusRejected:
		lg.Warn("message has been rejected by correspondant", zap.Any("message", msg))
	default:
		lg.Error("unexpected response processing status",
			zap.String("outcome", "message thrown away"),
			zap.Stringer("message_status", msg.MessageStatus),
			zap.Stringer("processing_status", msg.MessageStatus),
		)

		return
	}

	// update DB with next status: received
	//
	// NOTE: the update is not carried out if the status has not changed
	// (i.e if we process a redelivered version of a response already digested)
	msg.MessageStatus = repos.MessageStatusReceived
	msg.LastTime = time.Now().UTC()

	if err := p.rt.Repos().Messages().Update(ctx, msg); err != nil {
		if !errors.Is(err, repos.ErrAlreadyProcessed) {
			lg.Warn("could not update the message in DB",
				zap.String("outcome", "message NACK-ed will be redelivered"),
				zap.Error(err),
			)

			return
		}

		lg.Warn("message was already in received status",
			zap.String("outcome", "message confirmation redelivered"),
			zap.Error(err),
		)
	}

	// send back confirmation, so replays from consumer stop
	if err := p.confirmMessage(ctx, msg); err != nil {
		lg.Warn("message updated in DB as confirmed, but could not send confirmation",
			zap.String("outcome", "confirmation will be redelivered on the next round"),
			zap.Error(err),
		)
	}
}

// confirmMessage sends back a confirmation to the consumer.
func (p Producer) confirmMessage(parentCtx context.Context, msg repos.Message) error {
	_, span, lg := tracer.StartSpan(parentCtx, p)
	defer span.End()

	lg = lg.With(zap.String("id", msg.ID))

	// encode the message as []byte, using gob
	payload, err := msg.Bytes()
	if err != nil {
		lg.Warn("could not marshal message", zap.Error(err))

		return err
	}

	// send message with subject "postings.{ConsumerID}", don't expect any reply
	post := nats.NewMsg(p.publishedSubject(msg.ConsumerID))
	post.Data = payload
	post.Header.Set("trace_id", span.SpanContext().TraceID.String())
	post.Header.Set("span_id", span.SpanContext().SpanID.String())
	if err := p.nc.PublishMsg(post); err != nil {
		lg.Warn("could not publish confirmation",
			zap.String("outcome", "will be resubmitted later"),
			zap.Error(err),
		)
	}

	return nil
}

// replay messages to the consumer
func (p Producer) replay(parentCtx context.Context) error {
	spanCtx, span, lg := tracer.StartSpan(parentCtx, p)
	defer span.End()

	dbCtx, cancel := context.WithTimeout(spanCtx, p.Producer.MsgProcessTimeout)
	defer cancel()

	lg.Debug("looking for messages to be redelivered")

	notTooRecent := time.Now().UTC().Add(-1 * p.Producer.Replay.MinReplayDelay)
	iterator, err := p.rt.Repos().Messages().List(dbCtx, repos.MessagePredicate{
		FromProducer:     &p.ID,
		MaxMessageStatus: repos.NewMessageStatus(repos.MessageStatusReceived), // only replay messages which response has not been ack-ed yet
		Limit:            p.Producer.Replay.BatchSize,                         // only replay the oldest (batchsize) this time
		NotUpdatedSince:  &notTooRecent,                                       // check only not too recent
	})
	if err != nil {
		return err
	}
	defer func() {
		_ = iterator.Close()
	}()

	for iterator.Next() {
		msg, err := iterator.Item()
		if err != nil {
			return err
		}

		lg.Debug("found message to be redelivered", zap.Any("msg", msg))

		// republish message
		msg.ProducerReplays = 1 // will add 1
		msg.ConsumerReplays = 0 // will keep unchanged
		msg.LastTime = time.Now().UTC()

		if err = p.updateReplayAndSendMessage(dbCtx, msg); err != nil {
			return err
		}
	}

	return nil
}

func (p Producer) updateReplayAndSendMessage(parentCtx context.Context, msg repos.Message) error {
	spanCtx, span, lg := tracer.StartSpan(parentCtx, p)
	defer span.End()

	lg = lg.With(zap.String("id", msg.ID))

	// encode the message as []byte, using gob
	payload, err := msg.Bytes()
	if err != nil {
		lg.Warn("could not marshal message", zap.Error(err))

		return err
	}

	// write and commit to DB
	ctx, cancel := context.WithTimeout(spanCtx, p.Producer.MsgProcessTimeout)
	defer cancel()

	// force the update: no status is changed, just the replay counter
	if err := p.rt.Repos().Messages().UpdateReplay(ctx, msg); err != nil {
		lg.Warn("could not write replay counts in DB", zap.Error(err))

		return errors.Join(
			ErrProduceDB,
			err,
		)
	}
	lg.Debug("replay count updated")

	// send message with subject "postings.{ConsumerID}" and reply-to
	post := nats.NewMsg(p.publishedSubject(msg.ConsumerID))
	post.Data = payload
	post.Reply = p.subscribedSubject
	post.Header.Set("trace_id", span.SpanContext().TraceID.String())
	post.Header.Set("span_id", span.SpanContext().SpanID.String())
	if err := p.nc.PublishMsg(post); err != nil {
		lg.Warn("could not publish",
			zap.String("outcome", "will be resubmitted later"),
			zap.Error(err),
		)
	}

	lg.Debug("message redelivered", zap.String("consumer_id", msg.ConsumerID))

	return nil
}
