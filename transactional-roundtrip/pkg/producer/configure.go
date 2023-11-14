package producer

import (
	configkeys "github.com/fredbi/go-experiments/transactional-roundtrip/cmd/daemon/commands/config-keys"
	natsconfigkeys "github.com/fredbi/go-experiments/transactional-roundtrip/pkg/nats/config-keys"
	producerconfigkeys "github.com/fredbi/go-experiments/transactional-roundtrip/pkg/producer/config-keys"
	"github.com/spf13/viper"
)

// configSections resolves the config sections of interest or pick default settings.
func (p Producer) configSections() (*viper.Viper, *viper.Viper) {
	cfg := p.rt.Config()

	appConfig := cfg.Sub(configkeys.AppConfig)
	if appConfig == nil {
		appConfig = configkeys.DefaultAppConfig()
	}

	producerConfig := appConfig.Sub(configkeys.ProducerConfig)
	if producerConfig == nil {
		producerConfig = producerconfigkeys.DefaultProducerConfig()
	}

	natsConfig := cfg.Sub(configkeys.NatsConfig)
	if natsConfig == nil {
		natsConfig = natsconfigkeys.DefaultNATSConfig()
	}

	return natsConfig, producerConfig
}
