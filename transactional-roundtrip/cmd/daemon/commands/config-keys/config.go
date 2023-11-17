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

	// Participant ID used to route messages
	ParticipantID = "participantID"

	// Subsection consumer
	ConsumerConfig = "consumer"

	// Subsection producer
	ProducerConfig = "producer"
)

// nats section
const (
	NatsConfig = "nats"
)

// migrations section
const (
	MigrationsConfig = "migrations"
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
