package consumer

import (
	"bytes"
	"time"

	"github.com/fredbi/go-cli/config"
	natsembedded "github.com/fredbi/go-experiments/transactional-roundtrip/pkg/nats"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var defaultSettings = settings{
	Nats: natsembedded.DefaultSettings,
	Consumer: consumerSettings{
		Replay: replaySettings{
			BatchSize:      1000,
			WakeUp:         30 * time.Second,
			MinReplayDelay: 30 * time.Second,
		},
		MsgProcessTimeout: 5 * time.Second,
		ProcessTimeout:    30 * time.Second,
	},
}

type settings struct {
	Nats     natsembedded.Settings `yaml:"-" json:"-"`
	Consumer consumerSettings
}

type consumerSettings struct {
	Replay            replaySettings
	MsgProcessTimeout time.Duration
	ProcessTimeout    time.Duration
	NoReplay          bool
}

type replaySettings struct {
	BatchSize      uint64
	WakeUp         time.Duration
	MinReplayDelay time.Duration
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

// makeConfig resolves the config sections of interest or pick default settings.
func (p Consumer) makeConfig() (settings, error) {
	cfg := p.rt.Config()
	s := defaultSettings

	// global nats settings
	natsSettings, err := natsembedded.MakeSettings(cfg)
	if err != nil {
		return s, err
	}
	s.Nats = natsSettings

	appConfig := config.ViperSub(cfg, "app")
	if appConfig == nil {
		return s, nil
	}

	if err := appConfig.Unmarshal(&s); err != nil {
		return s, err
	}

	return s, nil
}
