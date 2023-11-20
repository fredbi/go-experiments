package producer

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
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

	if !p.Producer.NoReplay {
		<-latch
		grp.Go(p.startReplayer(ctx))
	} else {
		lg.Info("replayer is disabled")
	}

	grp.Go(func() error {
		<-latch
		return p.startHTTPHandler(ctx)
	})

	return grp.Wait()
}

func (p *Producer) Stop() error {
	if p.nc == nil {
		return nil
	}

	return p.nc.Drain()
}

func (p *Producer) startNatsClient(_ context.Context) (func(), error) {
	lg := p.rt.Logger().Bg().With(zap.String("operation", "start_nats_client"))

	p.publishedSubject = func(recipientID string) string {
		return fmt.Sprintf("%s.%s", p.Nats.Topics.Postings, recipientID)
	}
	p.subscribedSubject = fmt.Sprintf("%s.%s", p.Nats.Topics.Results, p.ID)

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
		zap.String("nats_url", p.Nats.URL),
		zap.String("subscribed_to_subject", p.subscribedSubject),
		zap.String("publish_to_subjects", p.publishedSubject("*")),
	)

	return clean, nil
}

// startReplayer starts a background replayer of un-acked messages
func (p Producer) startReplayer(ctx context.Context) func() error {
	lg := p.Logger().For(ctx).With(zap.String("operation", "replayer"))
	lg.Info("replay configuration",
		zap.Duration("replay_wakeup", p.Producer.Replay.WakeUp),
		zap.Uint64("replay_batch_size", p.Producer.Replay.BatchSize),
	)

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

	addr := ":" + strconv.Itoa(p.Producer.API.Port)
	lg.Info("listening on", zap.String("endpoint", addr))

	return http.ListenAndServe(addr, router) //#nosec
}
