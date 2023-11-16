package consumer

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"math/rand"
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
	ErrConsumeDB  = errors.New("could not write in DB")
	ErrConsumeMsg = errors.New("message taken into account, but service temporily unavailable")
)

type (
	Consumer struct {
		ID string
		rt injected.Runtime

		subscribedSubject string
		publishedSubject  func(string) string
		nc                *nats.Conn

		settings
	}
)

func New(rt injected.Runtime, id string) *Consumer {
	return &Consumer{
		ID: id,
		rt: rt,
	}
}

func (p Consumer) Logger() log.Factory {
	return p.rt.Logger()
}

// subscriptionHandler processes postings from producers.
//
// It updates the processing status then sends back a result.
//
// Until results are confirmed by the consumer, it will keep redelivering responses.
func (p Consumer) subscriptionHandler(incoming *nats.Msg) {
	spanCtx := natsembedded.SpanContextFromHeaders(context.Background(), incoming)
	parentCtx, cancel := context.WithTimeout(spanCtx, p.Consumer.MsgProcessTimeout)
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

	if err := msg.Validate(); err != nil {
		lg.Error("unexpected incoming message",
			zap.String("outcome", "message thrown away"),
			zap.Error(err),
		)

		return
	}

	// sanity check on processing status: expect only confirmations to have a decided outcome
	if msg.MessageStatus != repos.MessageStatusReceived && msg.ProcessingStatus != repos.ProcessingStatusPending {
		lg.Error("unexpected response status",
			zap.String("outcome", "message thrown away"),
			zap.Stringer("message_status", msg.MessageStatus),
			zap.Stringer("processing_status", msg.ProcessingStatus),
		)

		return
	}

	// sanity check on message status protocol
	switch msg.MessageStatus {
	case repos.MessageStatusNacked:
		// ok
		lg.Debug("nacked message received for processing", zap.Any("msg", msg))

		if err := p.process(ctx, &msg); err != nil {
			lg.Warn("message processing error",
				zap.String("outcome", "will be retried upon redelivery"),
				zap.Error(err),
			)

			return
		}

		// update the status of the message: now we can reply with a decided outcome
		msg.MessageStatus = repos.MessageStatusPosted
		msg.LastTime = time.Now().UTC()

		// encode the message as []byte, using gob
		payload, err := msg.Bytes()
		if err != nil {
			lg.Warn("could not marshal message",
				zap.String("outcome", "will be retried upon redelivery"),
				zap.Error(err),
			)

			return
		}

		// send result
		post := nats.NewMsg(p.publishedSubject(msg.ProducerID))
		post.Data = payload
		post.Reply = p.subscribedSubject
		post.Header.Add("trace_id", span.SpanContext().TraceID.String())
		post.Header.Add("span_id", span.SpanContext().SpanID.String())

		if err := p.nc.PublishMsg(post); err != nil {
			lg.Warn("could not publish",
				zap.String("outcome", "will be resubmitted later"),
				zap.Error(err),
			)
		}

		lg.Debug("result sent back to producer", zap.String("producer_id", msg.ProducerID), zap.Any("raw_msg", post))

	case repos.MessageStatusReceived:
		// ok. confirmation from producer: update private status
		lg.Debug("confirmation received from producer", zap.String("producer_id", msg.ProducerID), zap.Any("msg", msg))

		if err := p.rt.Repos().Messages().UpdateConfirmed(ctx, msg.ID, repos.MessageStatusConfirmed); err != nil {
			if errors.Is(err, repos.ErrAlreadyProcessed) {
				lg.Warn("duplicate message entry detected",
					zap.String("outcome", "no new message is created, user input is discarded"),
					zap.Error(err),
				)
			} else {
				lg.Warn("could not write in DB",
					zap.String("outcome", "will be retried upon redelivery"),
					zap.Error(err),
				)
			}
		}

	default:
		lg.Error("unexpected response status",
			zap.String("outcome", "message thrown away"),
			zap.Stringer("message_status", msg.MessageStatus),
			zap.Stringer("processing_status", msg.ProcessingStatus),
		)

		return
	}
}

//nolint:unparam
func (p Consumer) process(parentCtx context.Context, msg *repos.Message) error {
	_, cancel := context.WithTimeout(parentCtx, p.Consumer.ProcessTimeout)
	defer cancel()

	// TODO: some more realistic processing (e.g. build output files...)
	// for demo, just put random numbers to fill-in the balances
	msg.BalanceBefore = repos.NewDecimal(rand.Int63n(1_000_000_000), 2) //#nosec
	msg.BalanceAfter = repos.NewDecimal(rand.Int63n(1_000_000_000), 2)  //#nosec
	msg.ProcessingStatus = repos.ProcessingStatusOK

	return nil
}

// replay messages to the consumer
func (p Consumer) replay(ctx context.Context) error {
	dbCtx, cancel := context.WithTimeout(ctx, p.Consumer.MsgProcessTimeout)
	defer cancel()

	// list all currently unconfirmed messages and redeliver
	iterator, err := p.rt.Repos().Messages().List(dbCtx, repos.MessagePredicate{
		FromConsumer: &p.ID,
		Limit:        p.Consumer.Replay.BatchSize,
		Unconfirmed:  true,
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

		// republish message
		msg.ConsumerReplays++
		msg.LastTime = time.Now().UTC()

		if err = p.resendMessage(dbCtx, msg); err != nil {
			return err
		}
	}

	return nil
}

func (p Consumer) resendMessage(parentCtx context.Context, msg repos.Message) error {
	_, span, lg := tracer.StartSpan(parentCtx, p)
	defer span.End()

	lg = lg.With(zap.String("id", msg.ID))

	// encode the message as []byte, using gob
	payload, err := msg.Bytes()
	if err != nil {
		lg.Warn("could not marshal message", zap.Error(err))

		return err
	}

	// send message with subject "postings.{ProducerID}" and reply-to
	post := nats.NewMsg(p.publishedSubject(msg.ProducerID))
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

	return nil
}
