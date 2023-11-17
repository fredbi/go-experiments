package migrations

import (
	"bytes"
	"time"

	"github.com/fredbi/go-cli/config"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var defaultSettings = settings{
	Timeout:          30 * time.Second,
	MigrationTimeout: 5 * time.Second,
}

type settings struct {
	Timeout          time.Duration
	MigrationTimeout time.Duration
}

// DefaultSettings returns all defaults for this package as a viper register.
//
// This is primarily intended for documentation & help purpose.
func DefaultSettings() *viper.Viper {
	v := viper.New()
	v.SetConfigType("yaml")
	asYAML, _ := yaml.Marshal(defaultSettings)
	_ = v.ReadConfig(bytes.NewReader(asYAML))

	return v
}

func makeSettings(cfg *viper.Viper) (settings, error) {
	s := defaultSettings

	migrationsCfg := config.ViperSub(cfg, "migrations")
	if migrationsCfg == nil {
		return s, nil
	}

	if err := migrationsCfg.Unmarshal(&s); err != nil {
		return s, err
	}

	return s, nil
}
