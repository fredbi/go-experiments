package nats

import (
	"bytes"
	"time"

	"github.com/fredbi/go-cli/config"
	"github.com/nats-io/nats.go"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// DefaultSettings define defaults for the embedded NATS server and clients.
var DefaultSettings = Settings{
	URL: "nats://localhost:4222",
	Topics: TopicsSettings{
		Postings: "postings",
		Results:  "results",
	},
	Server: ServerSettings{
		MaxReconnect:           nats.DefaultOptions.MaxReconnect,
		ReconnectWait:          nats.DefaultOptions.ReconnectWait,
		StartupTimeout:         3 * time.Second,
		ClusterID:              "messaging",
		ClusterURL:             "nats://localhost:5333",
		ClusterRoutes:          "",
		ClusterHeadlessService: "",
	},
}

type (
	// Settings for embedded NATS.
	//
	// Primarily intended for being unmarshaled from a viper config.
	Settings struct {
		URL    string
		Topics TopicsSettings
		Server ServerSettings
	}

	TopicsSettings struct {
		Postings string
		Results  string
	}

	ServerSettings struct {
		StartupTimeout         time.Duration
		ReconnectWait          time.Duration
		MaxReconnect           int
		ClusterID              string
		ClusterURL             string
		ClusterRoutes          string
		ClusterHeadlessService string
		Debug                  ServerDebugSettings
	}

	ServerDebugSettings struct {
		Logs  bool
		Debug bool
		Trace bool
	}
)

// DefaultSettingsNATS returns all defaults for this package as a viper register.
//
// This is primarily intended for documentation & help purpose.
func DefaultSettingsNATS() *viper.Viper {
	v := viper.New()
	v.SetConfigType("yaml")
	asYAML, _ := yaml.Marshal(DefaultSettings)
	_ = v.ReadConfig(bytes.NewReader(asYAML))

	return v
}

func MakeSettings(cfg *viper.Viper) (Settings, error) {
	s := DefaultSettings
	if cfg == nil {
		return s, nil
	}

	cfg = config.ViperSub(cfg, "nats")
	if cfg == nil {
		return s, nil
	}

	if err := cfg.Unmarshal(&s); err != nil {
		return s, err
	}

	return s, nil
}
