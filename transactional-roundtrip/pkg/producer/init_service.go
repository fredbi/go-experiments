package producer

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	natsconfigkeys "github.com/fredbi/go-experiments/transactional-roundtrip/pkg/nats/config-keys"
	producerconfigkeys "github.com/fredbi/go-experiments/transactional-roundtrip/pkg/producer/config-keys"
	"github.com/nats-io/nats.go"
	"github.com/spf13/viper"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/plugin/ochttp/propagation/tracecontext"
	"go.opencensus.io/trace"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func (p *Producer) Start() error {
	natsConfig, producerConfig := p.configSections()
	cancellable, cancel := context.WithCancel(context.Background())
	defer cancel()

	cleanNats, err := p.startNatsClient(cancellable, natsConfig)
	if err != nil {
		cancel()

		return err
	}
	defer cleanNats()

	cleanReplayer, err := p.startReplayer(cancellable, producerConfig)
	if err != nil {
		cancel()

		return err
	}
	defer cleanReplayer()

	return p.startHTTPHandler(cancellable, producerConfig)
}

func (p *Producer) startNatsClient(_ context.Context, natsConfig *viper.Viper) (func(), error) {
	lg := p.rt.Logger().Bg().With(zap.String("operation", "start"))

	// settings for NATS
	natsURL := natsConfig.GetString(natsconfigkeys.URL)
	clusterID := natsConfig.GetString(natsconfigkeys.ClusterID)
	postingsTopic := natsConfig.GetString(natsconfigkeys.PostingsTopic)
	resultsTopic := natsConfig.GetString(natsconfigkeys.ResultsTopic)

	p.publishedSubject = func(recipientID string) string {
		return fmt.Sprintf("%s.%s", postingsTopic, recipientID)
	}
	p.subscribedSubject = fmt.Sprintf("%s.%s", resultsTopic, p.ID)

	// connect to the NATS cluster
	nc, err := nats.Connect(natsURL,
		nats.ReconnectWait(natsConfig.GetDuration(natsconfigkeys.ReconnectWait)),
		nats.MaxReconnects(natsConfig.GetInt(natsconfigkeys.MaxReconnect)),
	)
	if err != nil {
		return nil, err
	}

	p.nc = nc

	// subscribe to the the responses queue, with the ProducerID as subject
	subscription, err := nc.QueueSubscribe(p.subscribedSubject, p.ID, p.subscriptionHandler)
	if err != nil {
		return nil, err
	}

	clean := func() {
		_ = subscription.Unsubscribe()
	}

	lg.Info("producer connected to NATS cluster", zap.String("producer_id", p.ID), zap.String("url", natsURL), zap.String("cluster_id", clusterID))

	return clean, nil
}

//nolint:unparam
func (p *Producer) startReplayer(ctx context.Context, producerConfig *viper.Viper) (func(), error) {
	// Replays of unacked and messages in the background
	p.replayBatchSize = producerConfig.GetUint64(producerconfigkeys.ReplayBatchSize)
	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(p.replayer(groupCtx))

	return func() { _ = group.Wait() }, nil
}

// replayer starts a background replayer of un-acked messages
func (p Producer) replayer(ctx context.Context) func() error {
	lg := p.Logger().For(ctx).With(zap.String("operation", "http-handler"))

	return func() error {
		ticker := time.NewTicker(30 * time.Second) // TODO: settings
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

func (p *Producer) startHTTPHandler(ctx context.Context, producerConfig *viper.Viper) error {
	lg := p.rt.Logger().For(ctx).With(zap.String("module", "http-handler"))

	// TODO: serve other endpoints GET /message/{id}, GET /producer/{producerID}/messages?status=?
	// GET /producers GET /consumers GET /consumer/{consumerID}/messages?status=
	http.Handle("/message", ochttp.WithRouteTag(p, "/"))
	addr := ":" + producerConfig.GetString(producerconfigkeys.Port)
	p.jsonDecodeTimeout = producerConfig.GetDuration(producerconfigkeys.JSONDecodeTimeout)
	lg.Info("listening on", zap.String("endpoint", addr))

	return http.ListenAndServe(addr, nil) //#nosec
}

func (p *Producer) Stop() error {
	if p.nc == nil {
		return nil
	}

	return p.nc.Drain()
}

var p = &tracecontext.HTTPFormat{}

// spanContextFromHeaders extracts a trace span from the headers of a message.
func spanContextFromHeaders(parentCtx context.Context, msg *nats.Msg) context.Context {
	traceID := msg.Header.Get("trace_id")
	spanID := msg.Header.Get("span_id")
	spanCtx, ok := p.SpanContextFromHeaders(traceID, spanID)
	if !ok {
		return parentCtx
	}

	ctx, _ := trace.StartSpanWithRemoteParent(parentCtx, "incoming NATS message", spanCtx)

	return ctx
}
