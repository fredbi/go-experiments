package keys

import (
	"time"

	"github.com/spf13/viper"
)

// General settings
const (
	MessageHandlingTimeout = "timeout"
)

// HTTP API handler settings
const (
	Port              = "api.port"
	JSONDecodeTimeout = "api.jsonDecodeTimeout"
)

// Replay settings
const (
	ReplayBatchSize = "replay.batchSize"
	ReplayWakeUp    = "replay.wakeup"
)

func SetDefaults(cfg *viper.Viper) {
	cfg.SetDefault(Port, "9090")
	cfg.SetDefault(JSONDecodeTimeout, 5*time.Millisecond)
	cfg.SetDefault(ReplayBatchSize, 1000)
	cfg.SetDefault(ReplayWakeUp, 30*time.Second)
	cfg.SetDefault(MessageHandlingTimeout, 10*time.Second)
}

func DefaultProducerConfig() *viper.Viper {
	cfg := viper.New()
	SetDefaults(cfg)

	return cfg
}
