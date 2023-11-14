package keys

import (
	"time"

	"github.com/spf13/viper"
)

// General settings
const (
	MessageHandlingTimeout = "timeout"
)

// Replay settings
const (
	ReplayBatchSize = "replay.batchSize"
	ReplayWakeUp    = "replay.wakeup"
)

func SetDefaults(cfg *viper.Viper) {
	cfg.SetDefault(ReplayBatchSize, 1000)
	cfg.SetDefault(ReplayWakeUp, 30*time.Second)
	cfg.SetDefault(MessageHandlingTimeout, 10*time.Second)
}

func DefaultConsumerConfig() *viper.Viper {
	cfg := viper.New()
	SetDefaults(cfg)

	return cfg
}
