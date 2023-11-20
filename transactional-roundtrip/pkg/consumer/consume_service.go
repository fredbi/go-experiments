package consumer

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
	ErrConsumeDB               = errors.New("could not write in DB")
	ErrConsumeMsg              = errors.New("message taken into account, but service temporily unavailable")
	ErrUnexpectedMessageStatus = errors.New("unexpected message status")
)

type (
	Consumer struct {
		ID string
		rt injected.Runtime

		subscribedSubject string
		publishedSubject  func(string) string
		nc                *nats.Conn
		processor         MessageProcessor

		settings
	}
)

func New(rt injected.Runtime, id string) *Consumer {
	return &Consumer{
		ID:        id,
		rt:        rt,
		processor: NewDummyProcessor(rt),
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

	// unmarshal the incoming message
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

	if err := p.check(ctx, msg); err != nil {
		// send back a rejected result at once instead
		// of letting the producer retry infinitely.
		//
		// If this response doesn't get back to the producer, it wil be
		// replayed next time.
		lg.Warn("invalid message content caused rejection", zap.Any("msg", msg))

		msg.ProcessingStatus = repos.ProcessingStatusRejected
		cause := err.Error()
		msg.RejectionCause = &cause
		if err := p.sendMessage(ctx, msg); err != nil {
			return
		}

		return
	}

	// sanity check on message status protocol
	switch msg.MessageStatus {
	case repos.MessageStatusNacked:
		// ok
		lg.Debug("nacked message received for processing", zap.Any("msg", msg))

		if err := p.process(ctx, &msg); err != nil {
			return
		}

		if err := p.sendMessage(ctx, msg); err != nil {
			return
		}

		lg.Debug("result sent back to producer", zap.String("producer_id", msg.ProducerID), zap.Any("msg", msg))

	case repos.MessageStatusReceived:
		// ok. confirmation from producer: update private status
		lg.Debug("confirmation received from producer", zap.String("producer_id", msg.ProducerID), zap.Any("msg", msg))

		if err := p.confirmed(ctx, msg); err != nil {
			return
		}

	default:
		lg.Error("unexpected response status",
			zap.String("outcome", "message thrown away"),
			zap.String("subject", incoming.Subject),
			zap.String("from_producer", msg.ProducerID),
			zap.Stringer("message_status", msg.MessageStatus),
			zap.Stringer("processing_status", msg.ProcessingStatus),
		)

		return
	}
}

// check the incoming message.
//
// At this moment, we assume that invalid content is a permanent situation (e.g. fault in code).
//
// Transitory corruptions are assumed to be caught by the unmarshalling step.
func (p Consumer) check(parentCtx context.Context, msg repos.Message) error {
	_, span, lg := tracer.StartSpan(parentCtx, p)
	defer span.End()

	lg = lg.With(zap.String("id", msg.ID))

	if err := msg.Validate(); err != nil {
		lg.Error("invalid incoming message content",
			zap.String("outcome", "message thrown away"),
			zap.Error(err),
		)

		return err
	}

	// sanity check on processing status: expect only confirmations to have a decided outcome
	if msg.MessageStatus != repos.MessageStatusReceived && msg.ProcessingStatus != repos.ProcessingStatusPending {
		lg.Error("unexpected response status",
			zap.String("outcome", "message thrown away"),
			zap.Stringer("message_status", msg.MessageStatus),
			zap.Stringer("processing_status", msg.ProcessingStatus),
		)

		return ErrUnexpectedMessageStatus
	}

	return nil
}

func (p Consumer) process(parentCtx context.Context, msg *repos.Message) error {
	ctx, span, lg := tracer.StartSpan(parentCtx, p)
	defer span.End()

	lg = lg.With(zap.String("id", msg.ID))

	processCtx, cancel := context.WithTimeout(ctx, p.Consumer.ProcessTimeout)
	defer cancel()

	if err := p.processor.Process(processCtx, msg); err != nil {
		lg.Warn("message processing error",
			zap.String("outcome", "will be retried upon redelivery"),
			zap.Error(err),
		)

		return err
	}

	return nil
}

func (p Consumer) confirmed(parentCtx context.Context, msg repos.Message) error {
	ctx, span, lg := tracer.StartSpan(parentCtx, p)
	defer span.End()

	lg = lg.With(zap.String("id", msg.ID))

	if err := p.rt.Repos().Messages().UpdateConfirmed(ctx, msg.ID, repos.MessageStatusConfirmed); err != nil {
		if !errors.Is(err, repos.ErrAlreadyProcessed) {
			lg.Warn("could not write in DB",
				zap.String("outcome", "will be retried upon redelivery"),
				zap.Error(err),
			)

			return err
		}

		lg.Warn("duplicate message entry detected",
			zap.String("outcome", "no action"),
			zap.Error(err),
		)
	}

	return nil
}

// replay messages to the consumer
func (p Consumer) replay(parentCtx context.Context) error {
	spanCtx, span, lg := tracer.StartSpan(parentCtx, p)
	defer span.End()

	dbCtx, cancel := context.WithTimeout(spanCtx, p.Consumer.MsgProcessTimeout)
	defer cancel()

	lg.Debug("looking for messages to be redelivered")

	// list all currently unconfirmed messages and redeliver
	notTooRecent := time.Now().UTC().Add(-1 * p.Consumer.Replay.MinReplayDelay)
	iterator, err := p.rt.Repos().Messages().List(dbCtx, repos.MessagePredicate{
		FromConsumer:    &p.ID,
		Limit:           p.Consumer.Replay.BatchSize, // only replay the oldest (batchsize) this time
		Unconfirmed:     true,                        // check only unconfirmed messages
		NotUpdatedSince: &notTooRecent,               // check only not too recent
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
		if msg.ProcessingStatus == repos.ProcessingStatusPending {
			// process messages that could not be processed
			if err = p.process(dbCtx, &msg); err != nil {
				return err
			}
		}

		// update replay count
		msg.ProducerReplays = 0 // will keep unchanged
		msg.ConsumerReplays = 1 // will add 1
		if err = p.rt.Repos().Messages().UpdateReplay(dbCtx, msg); err != nil {
			lg.Warn("could not write replay counts in DB", zap.Error(err))

			return errors.Join(
				ErrConsumeDB,
				err,
			)
		}

		if err = p.sendMessage(dbCtx, msg); err != nil {
			return err
		}
	}

	return nil
}

// send back the message to the producer, with an updated timing
func (p Consumer) sendMessage(parentCtx context.Context, msg repos.Message) error {
	_, span, lg := tracer.StartSpan(parentCtx, p)
	defer span.End()

	lg = lg.With(zap.String("id", msg.ID))

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
			zap.String("outcome", "should be resubmitted later by the producer"),
			zap.Error(err),
		)
	}

	return nil
}
