package consumer

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	consumerconfigkeys "github.com/fredbi/go-experiments/transactional-roundtrip/pkg/consumer/config-keys"
	natsconfigkeys "github.com/fredbi/go-experiments/transactional-roundtrip/pkg/nats/config-keys"
	"github.com/nats-io/nats.go"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func (p *Consumer) Start() error {
	natsConfig, consumerConfig := p.configSections()
	p.msgProcessTimeout = consumerConfig.GetDuration(consumerconfigkeys.MessageHandlingTimeout)
	cancellable, cancel := context.WithCancel(context.Background())
	defer cancel()

	cleanNats, err := p.startNatsClient(cancellable, natsConfig)
	if err != nil {
		cancel()

		return err
	}
	defer cleanNats()

	waitReplayer, err := p.startReplayer(cancellable, consumerConfig)
	if err != nil {
		cancel()

		return err
	}

	return waitReplayer()
}

func (p *Consumer) startNatsClient(_ context.Context, natsConfig *viper.Viper) (func(), error) {
	lg := p.rt.Logger().Bg().With(zap.String("operation", "start"))

	// settings for NATS
	natsURL := natsConfig.GetString(natsconfigkeys.URL)
	clusterID := natsConfig.GetString(natsconfigkeys.ClusterID)
	postingsTopic := natsConfig.GetString(natsconfigkeys.PostingsTopic)
	resultsTopic := natsConfig.GetString(natsconfigkeys.ResultsTopic)

	p.publishedSubject = func(recipientID string) string {
		return fmt.Sprintf("%s.%s", resultsTopic, recipientID)
	}
	p.subscribedSubject = fmt.Sprintf("%s.%s", postingsTopic, p.ID)

	// connect to the NATS cluster
	nc, err := nats.Connect(natsURL,
		nats.ReconnectWait(natsConfig.GetDuration(natsconfigkeys.ReconnectWait)),
		nats.MaxReconnects(natsConfig.GetInt(natsconfigkeys.MaxReconnect)),
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

	lg.Info("consumer connected to NATS cluster", zap.String("consumer_id", p.ID), zap.String("url", natsURL), zap.String("cluster_id", clusterID))

	return clean, nil
}

//nolint:unparam
func (p *Consumer) startReplayer(ctx context.Context, consumerConfig *viper.Viper) (func() error, error) {
	// Replays of unacked and messages in the background
	p.replayBatchSize = consumerConfig.GetUint64(consumerconfigkeys.ReplayBatchSize)
	p.replayWakeUp = consumerConfig.GetDuration(consumerconfigkeys.ReplayWakeUp)

	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(p.replayer(groupCtx))

	return func() error { return group.Wait() }, nil
}

// replayer starts a background replayer of un-acked messages
func (p Consumer) replayer(ctx context.Context) func() error {
	lg := p.Logger().For(ctx).With(zap.String("operation", "http-handler"))

	return func() error {
		ticker := time.NewTicker(p.replayWakeUp)
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

				lg.Info("nothing to replay")

				break
			}
		}
	}
}

func (p *Consumer) Stop() error {
	if p.nc == nil {
		return nil
	}

	return p.nc.Drain()
}
