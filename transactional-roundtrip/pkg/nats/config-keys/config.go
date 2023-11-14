package keys

import (
	"time"

	nats "github.com/nats-io/nats.go"
	"github.com/spf13/viper"
)

const (
	URL            = "url"
	ClusterID      = "cluster_id"
	PostingsTopic  = "topics.postings"
	ResultsTopic   = "topics.results"
	StartupTimeout = "server.startupTimeout"
	MaxReconnect   = "client.maxReconnect"
	ReconnectWait  = "client.reconnectWait"
)

func SetDefaults(cfg *viper.Viper) {
	defaultNatsOptions := nats.DefaultOptions

	cfg.SetDefault(URL, nats.DefaultURL)
	cfg.SetDefault(ClusterID, "messaging")
	cfg.SetDefault(PostingsTopic, "postings")
	cfg.SetDefault(ResultsTopic, "results")

	// server settings
	cfg.SetDefault(StartupTimeout, 3*time.Second)

	// client settings
	cfg.SetDefault(MaxReconnect, defaultNatsOptions.MaxReconnect)
	cfg.SetDefault(ReconnectWait, defaultNatsOptions.ReconnectWait)

	/*
		AllowReconnect:     true,
		MaxReconnect:       DefaultMaxReconnect,
		ReconnectWait:      DefaultReconnectWait,
		ReconnectJitter:    DefaultReconnectJitter,
		ReconnectJitterTLS: DefaultReconnectJitterTLS,
		Timeout:            DefaultTimeout,
		PingInterval:       DefaultPingInterval,
		MaxPingsOut:        DefaultMaxPingOut,
		SubChanLen:         DefaultMaxChanLen,
		ReconnectBufSize:   DefaultReconnectBufSize,
		DrainTimeout:       DefaultDrainTimeout,
		FlusherTimeout:     DefaultFlusherTimeout,
	*/
}

func DefaultNATSConfig() *viper.Viper {
	cfg := viper.New()
	SetDefaults(cfg)

	return cfg
}
