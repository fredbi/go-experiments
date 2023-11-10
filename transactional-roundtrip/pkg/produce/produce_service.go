package produce

import (
	"bytes"
	"crypto/rand"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-nats/pkg/nats"
	"github.com/ThreeDotsLabs/watermill/message"
	configkeys "github.com/fredbi/go-experiments/transactional-roundtrip/cmd/daemon/commands/config-keys"
	"github.com/fredbi/go-experiments/transactional-roundtrip/pkg/injected"
	natsconfigkeys "github.com/fredbi/go-experiments/transactional-roundtrip/pkg/nats/config-keys"
	"github.com/fredbi/go-experiments/transactional-roundtrip/pkg/repo"
	"github.com/nats-io/stan.go"
	"github.com/oklog/ulid"
	"go.uber.org/zap"
)

type (
	Producer struct {
		ID              string
		rt              injected.Runtime
		publishedTopic  string
		subscribedTopic string
		publisher       message.Publisher
	}
)

func New(rt injected.Runtime, id string) *Producer {
	return &Producer{
		ID: id,
		rt: rt,
	}
}

func (p *Producer) Start() error {
	cfg := p.rt.Config()
	producerConfig := cfg.Sub(configkeys.ProducerConfig)
	if producerConfig == nil {
		return fmt.Errorf("missing configuration for producer. Expected a config key %s", configkeys.ProducerConfig)
	}
	natsConfig := cfg.Sub(configkeys.NatsConfig)
	if natsConfig == nil {
		return fmt.Errorf("missing configuration for NATS. Expected a config key %s", configkeys.NatsConfig)
	}

	natsURL := cfg.GetString(natsconfigkeys.URL)
	clusterID := cfg.GetString(natsconfigkeys.ClusterID)
	postingsTopic := cfg.GetString(natsconfigkeys.PostingsTopic)
	resultsTopic := cfg.GetString(natsconfigkeys.ResultsTopic)
	addr := ":" + cfg.GetString(natsconfigkeys.Port)

	publisher, err := nats.NewStreamingPublisher(
		nats.StreamingPublisherConfig{
			ClusterID: clusterID,
			ClientID:  p.ID,
			StanOptions: []stan.Option{
				stan.NatsURL(natsURL),
			},
			Marshaler: nats.GobMarshaler{}, // TODO: should probably just return the bytes slice
		},
		watermill.NewStdLogger(false, false), // TODO: inject my own tracable logger
	)
	if err != nil {
		return err
	}
	defer func() {
		_ = publisher.Close()
	}()

	// TODO: advertise to the cluster as a ProducerID

	p.publishedTopic = postingsTopic
	p.subscribedTopic = resultsTopic
	p.publisher = publisher

	http.Handle("/", p) // TODO: separate concerns

	lg := p.rt.Logger().Bg()
	lg.Info("Listening on ", zap.String("endpoint", addr))

	return http.ListenAndServe(addr, nil)
}

func (p Producer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(405)

		return
	}

	err := p.publish(w, r)
	if err != nil {
		lg := p.rt.Logger().For(r.Context())
		lg.Error("producing", zap.Error(err))
		w.WriteHeader(500)

		return
	}
}

func (p Producer) publish(w http.ResponseWriter, r *http.Request) error {
	var v struct {
		ConsumerID    string `json:"consumer_id"`
		OperationName string `json:"operation_name"`
	}
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&v); err != nil {
		return err
	}
	// TODO: check ConsumerID is legit

	ts := time.Now().UTC()
	uuid := ulid.MustNew(ulid.Timestamp(ts), rand.Reader)
	msg := repo.Message{
		ID:            uuid.String(),
		ProducerID:    p.ID,
		ConsumerID:    v.ConsumerID,
		InceptionTime: ts,
		LastTime:      ts,
		Payload: repo.Payload{
			OperationName: v.OperationName,
		},
	}
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	if err := enc.Encode(msg); err != nil {
		return err
	}

	post := message.NewMessage(uuid.String(), message.Payload(buffer.Bytes()))
	// TODO: write to DB

	if err := p.publisher.Publish(msg.ConsumerID, post); err != nil {
		return err
	}

	return nil
}

func (p Producer) Generate(n uint64) error {
	return nil // TODO
}
