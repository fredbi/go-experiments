package producer

import (
	"context"
	"crypto/rand"
	"errors"
	"net/http"
	"time"

	"github.com/fredbi/go-experiments/transactional-roundtrip/pkg/repos"
	"github.com/fredbi/go-trace/tracer"
	"github.com/go-chi/chi/v5"
	json "github.com/goccy/go-json"
	"github.com/oklog/ulid"
	"go.uber.org/zap"
)

// registerAPIRoutes adds all the http routes for the API handler
func (p Producer) registerAPIRoutes(router chi.Router) {
	router.Route("/", func(r chi.Router) {
		r.Post("/message", p.createMessage)
		r.Get("/message/{id}", p.getMessage)
		r.Get("/messages", p.listMessages)
	})
	router.Route("/probe", func(r chi.Router) {
		r.Get("/healthz", p.healthcheck)
		r.Get("/readyz", p.healthcheck)
	})
}

func (p Producer) createMessage(w http.ResponseWriter, r *http.Request) {
	parentCtx := r.Context()
	_, span, lg := tracer.StartSpan(parentCtx, p)
	defer span.End()

	if err := p.publishRequest(r); err != nil {
		lg.Warn("producing",
			zap.String("outcome", "user request is cancelled"),
			zap.Error(err),
		)

		if errors.Is(err, repos.ErrAlreadyProcessed) {
			w.WriteHeader(http.StatusUnprocessableEntity)

			return
		}

		w.WriteHeader(http.StatusInternalServerError)

		return
	}
}

func (p Producer) healthcheck(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("pass"))
}

func (p Producer) getMessage(w http.ResponseWriter, r *http.Request) {
	parentCtx := r.Context()
	ctx, span, lg := tracer.StartSpan(parentCtx, p)
	defer span.End()

	id := chi.URLParam(r, "id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	msg, err := p.rt.Repos().Messages().Get(ctx, id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)

		return
	}

	lg.Debug("retrieved", zap.String("id", id), zap.Any("message", msg))

	enc := json.NewEncoder(w)
	if err = enc.Encode(msg); err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		return
	}
}

func (p Producer) listMessages(w http.ResponseWriter, _ *http.Request) {
	// TODO: add list handler
	w.WriteHeader(http.StatusNotImplemented)
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
	lg.Debug("enriched user input", zap.Any("input_message", msg))

	if err := msg.Validate(); err != nil {
		return err
	}

	defer func() {
		lg.Debug("producing", zap.Any("message", msg))
	}()
	return p.createAndSendMessage(ctx, msg)
}

// decodeRequest attempts to decode an incoming JSON user input to build a Message prototype
func (p Producer) decodeRequest(parenCtx context.Context, r *http.Request) (repos.Message, error) {
	ctx, cancel := context.WithTimeout(parenCtx, p.Producer.API.JSONDecodeTimeout)
	defer cancel()

	var v repos.InputPayload
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.DecodeContext(ctx, &v); err != nil {
		return repos.Message{}, err
	}

	return v.AsMessage(), nil
}
