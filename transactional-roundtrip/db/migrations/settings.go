package migrations

import (
	"time"

	"github.com/fredbi/go-experiments/transactional-roundtrip/pkg/injected"
	"github.com/spf13/viper"
)

var defaultSettings = settings{
	Timeout:          30 * time.Second,
	MigrationTimeout: 5 * time.Second,
}

type settings struct {
	Timeout          time.Duration
	MigrationTimeout time.Duration
}

func makeSettings(cfg *viper.Viper) (settings, error) {
	s := defaultSettings

	migrationsCfg := injected.ViperSub(cfg, "migrations")
	if migrationsCfg == nil {
		return s, nil
	}

	if err := migrationsCfg.Unmarshal(&s); err != nil {
		return s, err
	}

	return s, nil
}
