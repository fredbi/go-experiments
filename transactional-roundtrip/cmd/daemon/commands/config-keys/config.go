package keys

import "github.com/spf13/viper"

const (
	LogLevel       = "log-level"
	DBConfig       = "databases"
	AppConfig      = "app"
	ConsumerConfig = "app.consumer"
	ProducerConfig = "app.producer"
	NatsConfig     = "nats"
	DSN            = "postgres.default.url"
)

// SetDefaults for CLI settings
func SetDefaults(cfg *viper.Viper) {
	cfg.SetDefault(AppConfig, map[string]interface{}{
		"consumer": map[string]interface{}{},
		"producer": map[string]interface{}{},
	})
	cfg.SetDefault(LogLevel, "info")
}
