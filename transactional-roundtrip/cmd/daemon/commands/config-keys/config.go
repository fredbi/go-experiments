package keys

import (
	natsconfigkeys "github.com/fredbi/go-experiments/transactional-roundtrip/pkg/nats/config-keys"
	producerconfigkeys "github.com/fredbi/go-experiments/transactional-roundtrip/pkg/producer/config-keys"
	pgrepoconfigkeys "github.com/fredbi/go-experiments/transactional-roundtrip/pkg/repos/pgrepo/config-keys"
	"github.com/spf13/viper"
)

// root section
const (
	LogLevel = "log.level"
)

// databases section
const (
	DBConfig = "databases"
)

// app section
const (
	AppConfig = "app"

	ParticipantID = "participant_id"

	// Subsection consumer
	ConsumerConfig = "consumer"

	// Subsection producer
	ProducerConfig = "producer"
)

// nats section
const (
	NatsConfig = "nats"
)

// SetDefaults for CLI settings
func SetDefaults(cfg *viper.Viper) {
	cfg.SetDefault(DBConfig, pgrepoconfigkeys.DefaultAllDBConfig())
	cfg.SetDefault(NatsConfig, natsconfigkeys.DefaultNATSConfig())
	cfg.SetDefault(AppConfig, DefaultAppConfig())
}

// DefaultConfig returns a default global configuration
func DefaultConfig() *viper.Viper {
	cfg := viper.New()
	SetDefaults(cfg)

	return cfg
}

// DefaultAppConfig returns a default "app" section.
func DefaultAppConfig() *viper.Viper {
	cfg := viper.New()
	cfg.SetDefault(AppConfig, map[string]interface{}{
		ConsumerConfig: producerconfigkeys.DefaultProducerConfig(),
		ProducerConfig: map[string]interface{}{}, // TODO
	})

	return cfg
}
