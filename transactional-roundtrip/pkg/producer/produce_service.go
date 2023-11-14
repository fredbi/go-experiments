package producer

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/gob"
	"errors"
	"net/http"
	"time"

	"github.com/fredbi/go-experiments/transactional-roundtrip/pkg/injected"
	"github.com/fredbi/go-experiments/transactional-roundtrip/pkg/repos"

	"github.com/fredbi/go-trace/log"
	"github.com/fredbi/go-trace/tracer"
	json "github.com/goccy/go-json"
	nats "github.com/nats-io/nats.go"
	"github.com/oklog/ulid"
	"go.uber.org/zap"
)

var (
	ErrProduceDB  = errors.New("could not write in DB")
	ErrProduceMsg = errors.New("message taken into account, but service temporily unavailable")
)

type (
	Producer struct {
		ID                string
		rt                injected.Runtime
		publishedSubject  func(string) string
		subscribedSubject string
		nc                *nats.Conn

		// settings
		replayBatchSize   uint64
		jsonDecodeTimeout time.Duration
	}
)

func New(rt injected.Runtime, id string) *Producer {
	return &Producer{
		ID: id,
		rt: rt,
	}
}

func (p Producer) Logger() log.Factory {
	return p.rt.Logger()
}

func (p Producer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: separate HTTP handlers in a different object
	if r.Method != "POST" {
		w.WriteHeader(405)

		return
	}

	err := p.publishRequest(r)
	if err != nil {
		lg := p.rt.Logger().For(r.Context())
		lg.Warn("producing",
			zap.String("outcome", "user request is cancelled"),
			zap.Error(err),
		)

		w.WriteHeader(500)

		return
	}
}

// publishRequest takes a user input from the HTTP API, records this message then forwards this to the appropriate consumer.
func (p Producer) publishRequest(r *http.Request) error {
	parentCtx := r.Context()
	ctx, span, lg := tracer.StartSpan(parentCtx, p)
	defer span.End()

	// decode the incoming JSON from the HTTP request
	msg, err := p.decodeRequest(ctx, r)
	if err != nil {
		lg.Warn("request body could not be encoded",
			zap.String("outcome", "incoming message is rejected"),
			zap.Error(err),
		)

		return err
	}

	// determine a new UUID-like unique ID for this message
	ts := time.Now().UTC()
	uuid := ulid.MustNew(ulid.Timestamp(ts), rand.Reader)

	msg.ID = uuid.String()
	msg.ProducerID = p.ID
	msg.InceptionTime = ts
	msg.LastTime = ts
	msg.MessageStatus = repos.MessageStatusNacked
	msg.ProcessingStatus = repos.ProcessingStatusPending

	if err := msg.Validate(); err != nil {
		return err
	}

	return p.createAndSendMessage(ctx, msg)
}

func (p Producer) createAndSendMessage(parentCtx context.Context, msg repos.Message) error {
	spanCtx, span, lg := tracer.StartSpan(parentCtx, p)
	defer span.End()

	ctx, cancel := context.WithCancel(spanCtx)
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

	// send message with subject "postings.{ConsumerID}" and reply-to
	post := nats.NewMsg(p.publishedSubject(msg.ConsumerID))
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

	return nil
}

// decodeRequest attempts to decode an incoming JSON user input to build a Message prototype
func (p Producer) decodeRequest(parenCtx context.Context, r *http.Request) (repos.Message, error) {
	ctx, cancel := context.WithTimeout(parenCtx, p.jsonDecodeTimeout)
	defer cancel()

	var v repos.InputPayload
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.DecodeContext(ctx, &v); err != nil {
		return repos.Message{}, err
	}

	return v.AsMessage(), nil
}

// subscriptionHandler processes responses from consumers.
//
// It updates the processing status then sends back a confirmation.
//
// Until confirmations are acknowledged by the consumer, it will keep redelivering responses.
func (p Producer) subscriptionHandler(incoming *nats.Msg) {
	spanCtx := spanContextFromHeaders(context.Background(), incoming)
	parentCtx, cancel := context.WithTimeout(spanCtx, 10*time.Second) // TODO: configurable producer settting
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

	// sanity check on message status protocol
	if msg.MessageStatus != repos.MessageStatusReceived {
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
		lg.Info("message has been rejected by correspondant")
	default:
		lg.Error("unexpected response processing status",
			zap.String("outcome", "message thrown away"),
			zap.String("outcome", "message thrown away"),
			zap.Stringer("message_status", msg.MessageStatus),
			zap.Stringer("processing_status", msg.MessageStatus),
		)

		return
	}

	// update DB with next status: received
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

		lg.Info("message was already in received status",
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
	msg.MessageStatus = repos.MessageStatusConfirmed
	msg.LastTime = time.Now().UTC()
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
func (p Producer) replay(ctx context.Context) error {
	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second) // TODO: producer config
	defer cancel()

	iterator, err := p.rt.Repos().Messages().List(dbCtx, repos.MessagePredicate{
		FromProducer:     &p.ID,
		MaxMessageStatus: repos.NewMessageStatus(repos.MessageStatusReceived),
		Limit:            p.replayBatchSize,
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
		msg.ProducerReplays++
		msg.LastTime = time.Now().UTC()

		if err = p.updateAndSendMessage(dbCtx, msg); err != nil {
			return err
		}
	}

	return nil
}

func (p Producer) updateAndSendMessage(parentCtx context.Context, msg repos.Message) error {
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
	ctx, cancel := context.WithTimeout(spanCtx, 5*time.Second)
	defer cancel()
	if err := p.rt.Repos().Messages().Update(ctx, msg); err != nil {
		if errors.Is(err, repos.ErrAlreadyProcessed) {
			lg.Warn("duplicate update entry detected",
				zap.String("outcome", "no update, no need to redeliver"),
				zap.Error(err),
			)

			return err
		}

		lg.Warn("could not write in DB, not accepted in the system", zap.Error(err))

		return errors.Join(
			ErrProduceDB,
			err,
		)
	}

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

	return nil
}
