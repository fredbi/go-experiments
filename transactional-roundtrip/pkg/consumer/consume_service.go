package consumer

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/gob"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/cockroachdb/apd"
	configkeys "github.com/fredbi/go-experiments/transactional-roundtrip/cmd/daemon/commands/config-keys"
	"github.com/fredbi/go-experiments/transactional-roundtrip/pkg/injected"
	natsconfigkeys "github.com/fredbi/go-experiments/transactional-roundtrip/pkg/nats/config-keys"
	"github.com/fredbi/go-experiments/transactional-roundtrip/pkg/repos"
	"github.com/fredbi/go-trace/log"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

var (
	ErrConsumeDB  = errors.New("consumer could not write in DB")
	ErrConsumeMsg = errors.New("consumer message taken into account, but service temporily unavailable")
)

type Consumer struct {
	ID                string
	rt                injected.Runtime
	nc                *nats.Conn
	publishedSubject  func(string) string
	subscribedSubject string
}

func New(rt injected.Runtime, id string) *Consumer {
	return &Consumer{
		ID: id,
		rt: rt,
	}
}

func (p Consumer) Logger() log.Factory {
	return p.rt.Logger()
}

func (p Consumer) Stop() error {
	if p.nc == nil {
		return nil
	}

	return p.nc.Drain()
}
func (p *Consumer) Start() error {
	cfg := p.rt.Config()
	lg := p.Logger().Bg().With(zap.String("operation", "start"))

	appConfig := cfg.Sub(configkeys.AppConfig)
	if appConfig == nil {
		return fmt.Errorf("missing configuration for app. Expected a config key %s", configkeys.AppConfig)
	}
	consumerConfig := appConfig.Sub(configkeys.ConsumerConfig)
	if consumerConfig == nil {
		return fmt.Errorf("missing configuration for consumer. Expected a config key %s", configkeys.ConsumerConfig)
	}
	natsConfig := cfg.Sub(configkeys.NatsConfig)
	if natsConfig == nil {
		natsConfig = natsconfigkeys.DefaultNATSConfig()
	}

	natsURL := natsConfig.GetString(natsconfigkeys.URL)
	clusterID := natsConfig.GetString(natsconfigkeys.ClusterID)
	postingsTopic := natsConfig.GetString(natsconfigkeys.PostingsTopic)
	resultsTopic := natsConfig.GetString(natsconfigkeys.ResultsTopic)
	p.publishedSubject = func(recipientID string) string {
		return fmt.Sprintf("%s.%s", resultsTopic, recipientID)
	}
	p.subscribedSubject = fmt.Sprintf("%s.%s", postingsTopic, p.ID)

	// 1. Connects to the NATS cluster
	nc, err := nats.Connect(natsURL,
		nats.ReconnectWait(time.Second),
		nats.MaxReconnects(60),
	)
	if err != nil {
		return err
	}
	defer func() {
		_ = nc.Drain()
	}()

	p.nc = nc

	// 2. Subscribe to the the responses queue, with the ProducerID as subject
	subscription, err := nc.QueueSubscribe(p.subscribedSubject, p.ID, p.subscriptionHandler)
	if err != nil {
		return err
	}
	defer func() {
		_ = subscription.Unsubscribe()
	}()

	// 3. Run a replay of unconfirmed messages in the background
	cancellable, cancel := context.WithCancel(context.Background())
	group, groupCtx := errgroup.WithContext(cancellable)
	group.Go(p.replayer(groupCtx))

	defer func() {
		_ = group.Wait()
	}()
	defer cancel()

	lg.Info("producer connected to NATS cluster", zap.String("producer_id", p.ID), zap.String("url", natsURL), zap.String("cluster_id", clusterID))
	return nil
}

func (p *Consumer) subscriptionHandler(incoming *nats.Msg) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// TODO: trace span
	lg := p.Logger().For(ctx)
	replyTo := incoming.Reply

	buffer := bytes.NewReader(incoming.Data)
	var msg repos.Message
	dec := gob.NewDecoder(buffer)
	if err := dec.Decode(&msg); err != nil {
		lg.Warn("invalid message content. Message needs to be redelivered", zap.Error(err))

		return
	}

	if msg.MessageStatus != repos.MessageStatusConfirmed {
		// process message (skipped when this is a mere confirmation)
		if err := p.process(ctx, msg); err != nil {
			lg.Warn("message processing error. Status rejected", zap.Error(err))

			msg.ProcessingStatus = repos.ProcessingStatusRejected
			cause := err.Error()
			msg.RejectionCause = &cause
		} else {
			msg.ProcessingStatus = repos.ProcessingStatusOK
			balanceBefore, balanceAfter := rand.Int63(), rand.Int63() //#nosec
			msg.BalanceBefore = apd.New(balanceBefore, 100)
			msg.BalanceAfter = apd.New(balanceAfter, 100)
		}

		// update DB with response status received
		msg.MessageStatus = repos.MessageStatusPosted
	}

	msg.LastTime = time.Now().UTC()
	if err := p.updateAndSendMessage(ctx, msg, replyTo); err != nil {
		lg.Warn("could not update. Message needs to be redelivered", zap.Error(err))
	}
}

func (p Consumer) process(_ context.Context, _ repos.Message) error {
	return nil // TODO
}

// replayer starts a background replayer of un-acked messages
func (p Consumer) replayer(ctx context.Context) func() error {
	lg := p.Logger().For(ctx)

	return func() error {
		// TODO: span trace
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-ticker.C:
				// wake up the replay
				if err := p.replay(ctx); err != nil {
					if !errors.Is(err, sql.ErrNoRows) {
						lg.Warn("could not replay messages. Will try again next time", zap.Error(err))
					} else {
						lg.Info("nothing to replay")
					}

					break
				}
			}
		}
	}
}

func (p Consumer) replay(ctx context.Context) error {
	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	iterator, err := p.rt.Repos().Messages().List(dbCtx, repos.MessagePredicate{
		FromConsumer:     &p.ID,
		MaxMessageStatus: repos.NewMessageStatus(repos.MessageStatusConfirmed),
		// TODO: limit
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

		if err = p.updateAndSendMessage(dbCtx, msg, ""); err != nil {
			return err
		}
	}

	return nil
}

func (p Consumer) updateAndSendMessage(parentCtx context.Context, msg repos.Message, replyTo string) error {
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	lg := p.rt.Logger().For(ctx)

	// 3. Encode the message as []byte, using gob
	payload, err := msg.Bytes()
	if err != nil {
		lg.Warn("could not marshal message", zap.Error(err))

		return err
	}

	// 4. write and commit to DB
	if err := p.rt.Repos().Messages().Update(ctx, msg); err != nil {
		lg.Warn("could not write in DB, not accepted in the system", zap.Error(err))

		return errors.Join(
			ErrConsumeDB,
			err,
		)
	}

	if msg.MessageStatus == repos.MessageStatusConfirmed {
		return nil
	}

	// 5. Send message with subject "results.{ProducerID}" and reply-to
	if replyTo == "" {
		replyTo = p.publishedSubject(msg.ProducerID)
	}
	post := nats.NewMsg(replyTo)
	post.Subject = replyTo
	post.Data = payload
	post.Reply = p.subscribedSubject
	if err := p.nc.PublishMsg(post); err != nil {
		lg.Warn("could not publish. Will be resubmitted later", zap.Error(err))
	}

	return nil
}
