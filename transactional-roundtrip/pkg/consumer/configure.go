package consumer

import (
	configkeys "github.com/fredbi/go-experiments/transactional-roundtrip/cmd/daemon/commands/config-keys"
	consumerconfigkeys "github.com/fredbi/go-experiments/transactional-roundtrip/pkg/consumer/config-keys"
	natsconfigkeys "github.com/fredbi/go-experiments/transactional-roundtrip/pkg/nats/config-keys"
	"github.com/spf13/viper"
)

// configSections resolves the config sections of interest or pick default settings.
func (p Consumer) configSections() (*viper.Viper, *viper.Viper) {
	cfg := p.rt.Config()

	appConfig := cfg.Sub(configkeys.AppConfig)
	if appConfig == nil {
		appConfig = configkeys.DefaultAppConfig()
	}

	consumerConfig := appConfig.Sub(configkeys.ConsumerConfig)
	if consumerConfig == nil {
		consumerConfig = consumerconfigkeys.DefaultConsumerConfig()
	}

	natsConfig := cfg.Sub(configkeys.NatsConfig)
	if natsConfig == nil {
		natsConfig = natsconfigkeys.DefaultNATSConfig()
	}

	return natsConfig, consumerConfig
}
