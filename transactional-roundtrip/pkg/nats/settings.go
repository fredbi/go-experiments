package nats

import (
	"time"

	"github.com/nats-io/nats.go"
	"github.com/spf13/viper"
)

// DefaultSettings define defaults for the embedded NATS server and clients.
var DefaultSettings = Settings{
	URL:       nats.DefaultURL,
	ClusterID: "messaging",
	Postings:  "postings",
	Results:   "results",
	Server: ServerSettings{
		MaxReconnect:   nats.DefaultOptions.MaxReconnect,
		ReconnectWait:  nats.DefaultOptions.ReconnectWait,
		StartupTimeout: 3 * time.Second,
	},
}

type (
	// Settings for embedded NATS.
	//
	// Primarily intended for being unmarshaled from a viper config.
	Settings struct {
		URL       string
		ClusterID string
		Postings  string
		Results   string
		Server    ServerSettings
	}

	ServerSettings struct {
		StartupTimeout time.Duration
		ReconnectWait  time.Duration
		MaxReconnect   int
	}
)

func MakeSettings(cfg *viper.Viper) (Settings, error) {
	s := DefaultSettings

	natsConfig := cfg.Sub("nats")
	if natsConfig == nil {
		return s, nil
	}

	if err := natsConfig.Unmarshal(&s); err != nil {
		return s, err
	}

	return s, nil
}
