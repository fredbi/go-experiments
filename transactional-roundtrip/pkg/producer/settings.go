package producer

import (
	"time"

	configkeys "github.com/fredbi/go-experiments/transactional-roundtrip/cmd/daemon/commands/config-keys"
	natsembedded "github.com/fredbi/go-experiments/transactional-roundtrip/pkg/nats"
)

var defaultSettings = settings{
	Nats: natsembedded.DefaultSettings,
	Producer: producerSettings{
		Replay: replaySettings{
			BatchSize: 1000,
			WakeUp:    30 * time.Second,
		},
		API: apiSettings{
			Port: "9990",
		},
		MsgProcessTimeout: 5 * time.Second,
	},
}

type (
	settings struct {
		Nats     natsembedded.Settings
		Producer producerSettings
	}

	producerSettings struct {
		Replay            replaySettings
		API               apiSettings
		MsgProcessTimeout time.Duration
	}

	replaySettings struct {
		BatchSize uint64
		WakeUp    time.Duration
	}

	apiSettings struct {
		Port              string
		JSONDecodeTimeout time.Duration
	}
)

func (p Producer) makeConfig() (settings, error) {
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
