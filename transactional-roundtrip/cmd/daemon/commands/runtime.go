package commands

import (
	"time"

	"github.com/fredbi/go-cli/cli/cli-utils/resolve"
	configkeys "github.com/fredbi/go-experiments/transactional-roundtrip/cmd/daemon/commands/config-keys"
	"github.com/fredbi/go-experiments/transactional-roundtrip/pkg/injected"
	"github.com/fredbi/go-experiments/transactional-roundtrip/pkg/repos"
	"github.com/fredbi/go-experiments/transactional-roundtrip/pkg/repos/pgrepo"
	"github.com/fredbi/go-trace/log"
	"github.com/goombaio/namegenerator"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type runtime struct {
	id     string
	logger log.Factory
	db     *sqlx.DB
	cfg    *viper.Viper
	repos  repos.Iface
	closer func() error
}

var _ injected.Runtime = runtime{}

func (r runtime) DB() *sqlx.DB {
	return r.db
}

func (r runtime) Logger() log.Factory {
	return r.logger
}

func (r runtime) Config() *viper.Viper {
	return r.cfg
}

func (r runtime) Repos() repos.Iface {
	return r.repos
}

func (r runtime) ID() string {
	return r.id
}

func (r runtime) Close() error {
	return r.closer()
}

func newRuntimeForCommand(c *cobra.Command) (runtime, error) {
	_, zlg, cfg := resolve.InjectedZapConfig(c,
		resolve.WithZapLoggerDefaulter(zap.NewNop),
		resolve.WithConfigDefaulter(configkeys.DefaultConfig),
	)
	name := c.Name()

	// determine the participant ID.
	// Several instances of a server may share the same ID.
	participantID := cfg.GetString(inSection(configkeys.AppConfig, configkeys.ParticipantID))
	if participantID == "" {
		namer := namegenerator.NewNameGenerator(time.Now().UTC().UnixNano())
		participantID = namer.Generate()
	}

	lg := log.NewFactory(zlg.Named(name)).With(zap.String("participant_id", participantID))
	rt := runtime{
		id:     participantID,
		logger: lg,
		cfg:    cfg,
	}

	// open DB
	r := pgrepo.New(name, lg, cfg)
	if err := r.Start(); err != nil {
		return rt, err
	}

	rt.repos = r
	rt.db = r.DB()
	rt.closer = rt.db.Close

	return rt, nil
}
