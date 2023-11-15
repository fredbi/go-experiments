package producer

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	logmiddleware "github.com/fredbi/go-trace/log/middleware"
	"github.com/fredbi/go-trace/tracer"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func (p *Producer) Start() error {
	settings, err := p.makeConfig()
	if err != nil {
		return err
	}

	p.settings = settings
	cancellable, cancel := context.WithCancel(context.Background())
	defer cancel()

	cleanNats, err := p.startNatsClient(cancellable)
	if err != nil {
		cancel()

		return err
	}
	defer cleanNats()

	cleanReplayer, err := p.startReplayer(cancellable)
	if err != nil {
		cancel()

		return err
	}
	defer cleanReplayer()

	return p.startHTTPHandler(cancellable)
}

func (p *Producer) startNatsClient(_ context.Context) (func(), error) {
	lg := p.rt.Logger().Bg().With(zap.String("operation", "start"))

	p.publishedSubject = func(recipientID string) string {
		return fmt.Sprintf("%s.%s", p.Nats.Postings, recipientID)
	}
	p.subscribedSubject = fmt.Sprintf("%s.%s", p.Nats.Results, p.ID)

	// connect to the NATS cluster
	nc, err := nats.Connect(p.Nats.URL,
		nats.ReconnectWait(p.Nats.Server.ReconnectWait),
		nats.MaxReconnects(p.Nats.Server.MaxReconnect),
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

	lg.Info("producer connected to NATS cluster",
		zap.String("producer_id", p.ID),
		zap.String("url", p.Nats.URL),
		zap.String("cluster_id", p.Nats.ClusterID),
	)

	return clean, nil
}

//nolint:unparam
func (p *Producer) startReplayer(ctx context.Context) (func(), error) {
	// replays unacked and messages in the background
	lg := p.rt.Logger().Bg()
	lg.Info("replay configuration",
		zap.Duration("replay_wakeup", p.Producer.Replay.WakeUp),
		zap.Uint64("replay_batch_size", p.Producer.Replay.BatchSize),
	)

	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(p.replayer(groupCtx))

	return func() { _ = group.Wait() }, nil
}

// replayer starts a background replayer of un-acked messages
func (p Producer) replayer(ctx context.Context) func() error {
	lg := p.Logger().For(ctx).With(zap.String("operation", "http-handler"))

	return func() error {
		ticker := time.NewTicker(p.Producer.Replay.WakeUp)
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

func (p *Producer) startHTTPHandler(ctx context.Context) error {
	lg := p.rt.Logger().For(ctx).With(zap.String("operation", "start_http_handler"))

	router := chi.NewRouter()
	router.Use(middleware.Recoverer)
	router.Use(tracer.Middleware())
	router.Use(logmiddleware.LogRequests(p.rt.Logger()))
	router.Route("/", func(r chi.Router) {
		r.Post("/message", p.createMessage)
		r.Get("/message/{id}", p.getMessage)
		r.Get("/messages", p.listMessages)
	})

	addr := ":" + p.Producer.API.Port
	lg.Info("listening on", zap.String("endpoint", addr))

	return http.ListenAndServe(addr, router) //#nosec
}

func (p *Producer) Stop() error {
	if p.nc == nil {
		return nil
	}

	return p.nc.Drain()
}
