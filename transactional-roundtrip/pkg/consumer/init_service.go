package consumer

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func (p *Consumer) Start() error {
	s, err := p.makeConfig()
	if err != nil {
		return err
	}

	p.settings = s
	grp, ctx := errgroup.WithContext(context.Background())
	latch := make(chan struct{})
	lg := p.rt.Logger().Bg().With(zap.String("operation", "start_consumer"))

	grp.Go(func() error {
		cleanNats, err := p.startNatsClient(ctx)
		if err != nil {
			return err
		}
		defer cleanNats()

		close(latch)
		<-ctx.Done()

		return ctx.Err()
	})

	if !p.Consumer.NoReplay {
		<-latch
		grp.Go(p.startReplayer(ctx))
	} else {
		lg.Info("replayer is disabled")
	}

	return grp.Wait()
}

func (p *Consumer) Stop() error {
	if p.nc == nil {
		return nil
	}

	return p.nc.Drain()
}

func (p *Consumer) startNatsClient(_ context.Context) (func(), error) {
	lg := p.rt.Logger().Bg().With(zap.String("operation", "start_nats_client"))

	p.publishedSubject = func(recipientID string) string {
		return fmt.Sprintf("%s.%s", p.Nats.Topics.Results, recipientID)
	}
	p.subscribedSubject = fmt.Sprintf("%s.%s", p.Nats.Topics.Postings, p.ID)

	// connect to the NATS cluster
	nc, err := nats.Connect(p.Nats.URL,
		nats.ReconnectWait(p.Nats.Server.ReconnectWait),
		nats.MaxReconnects(p.Nats.Server.MaxReconnect),
	)
	if err != nil {
		return nil, err
	}

	p.nc = nc

	// subscribe to the the responses queue, with the ConsumerID as subject
	subscription, err := nc.QueueSubscribe(p.subscribedSubject, p.ID, p.subscriptionHandler)
	if err != nil {
		return nil, err
	}

	clean := func() {
		_ = subscription.Unsubscribe()
	}

	lg.Info("consumer connected to NATS cluster as a NATS client",
		zap.String("consumer_id", p.ID),
		zap.String("nats_url", p.Nats.URL),
		zap.String("subscribed_to_subject", p.subscribedSubject),
		zap.String("publish_to_subjects", p.publishedSubject("*")),
	)

	return clean, nil
}

// replayer starts a background replayer of un-acked messages
func (p Consumer) startReplayer(ctx context.Context) func() error {
	lg := p.Logger().For(ctx).With(zap.String("operation", "consumer_replayer"))
	lg.Info("replay configuration",
		zap.Duration("replay_wakeup", p.Consumer.Replay.WakeUp),
		zap.Uint64("replay_batch_size", p.Consumer.Replay.BatchSize),
	)

	return func() error {
		ticker := time.NewTicker(p.Consumer.Replay.WakeUp)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-ticker.C:
				// wake up the replay
				err := p.replay(ctx)
				if err == nil {
					break
				}

				if !errors.Is(err, sql.ErrNoRows) {
					lg.Warn("could not replay messages",
						zap.String("outcome", "will try again next time"),
						zap.Error(err),
					)

					break
				}

				lg.Debug("nothing to replay")

				break
			}
		}
	}
}
