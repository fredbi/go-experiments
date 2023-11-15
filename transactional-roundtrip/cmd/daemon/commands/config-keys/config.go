package keys

import (
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

// SetDefaults for CLI settings.
//
// At this moment, all defaults are either defined from CLI flag bindings
// or from the modules owners of their subsection.
func SetDefaults(_ *viper.Viper) {
}

// DefaultConfig returns a default global configuration
func DefaultConfig() *viper.Viper {
	cfg := viper.New()
	SetDefaults(cfg)

	return cfg
}
