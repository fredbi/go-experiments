package keys

import (
	"github.com/spf13/viper"
)

const ()

func SetDefaults(_ *viper.Viper) {
}

func DefaultConsumerConfig() *viper.Viper {
	cfg := viper.New()
	SetDefaults(cfg)

	return cfg
}
