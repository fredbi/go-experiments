package keys

import (
	"time"

	"github.com/spf13/viper"
)

// HTTP handler settings
const (
	Port              = "port"
	JSONDecodeTimeout = "jsonDecodeTimeout"
)

// Replay settings
const (
	ReplayBatchSize = "replayBatchSize"
)

func SetDefaults(cfg *viper.Viper) {
	cfg.SetDefault(Port, "9090")
	cfg.SetDefault(JSONDecodeTimeout, 5*time.Millisecond)
	cfg.SetDefault(ReplayBatchSize, 1000)
}

func DefaultProducerConfig() *viper.Viper {
	cfg := viper.New()
	SetDefaults(cfg)

	return cfg
}
