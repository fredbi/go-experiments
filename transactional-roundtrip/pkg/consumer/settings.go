package consumer

import (
	"time"

	configkeys "github.com/fredbi/go-experiments/transactional-roundtrip/cmd/daemon/commands/config-keys"
	natsembedded "github.com/fredbi/go-experiments/transactional-roundtrip/pkg/nats"
)

var defaultSettings = settings{
	Nats: natsembedded.DefaultSettings,
	Consumer: consumerSettings{
		Replay: replaySettings{
			BatchSize: 1000,
			WakeUp:    30 * time.Second,
		},
		MsgProcessTimeout: 5 * time.Second,
		ProcessTimeout:    30 * time.Second,
	},
}

type settings struct {
	Nats     natsembedded.Settings
	Consumer consumerSettings
}

type consumerSettings struct {
	Replay            replaySettings
	MsgProcessTimeout time.Duration
	ProcessTimeout    time.Duration
}

type replaySettings struct {
	BatchSize uint64
	WakeUp    time.Duration
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

	appConfig := cfg.Sub(configkeys.AppConfig)
	if appConfig == nil {
		return s, nil
	}

	if err := appConfig.Unmarshal(&s); err != nil {
		return s, err
	}

	return s, nil
}
