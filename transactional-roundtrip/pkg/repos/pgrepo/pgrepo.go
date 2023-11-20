package pgrepo

import (
	"github.com/fredbi/go-experiments/transactional-roundtrip/pkg/repos"
	messages "github.com/fredbi/go-experiments/transactional-roundtrip/pkg/repos/pg-messages"
	"github.com/fredbi/go-trace/log"
	pgpool "github.com/fredbi/pgxutils/pgrepo"
	"github.com/spf13/viper"
)

var _ repos.Iface = &Repository{}

// Repository knows how to handle a postgres backend database.
//
// The database driver is instrumented for tracing.
type Repository struct {
	*pgpool.Repository
	messageRepo repos.MessageRepo
	cfg         *viper.Viper
}

// New new postgres repository serving Messages.
//
// The new repository needs to be started wih Start() in order to create the connection pools.
func New(app string, lg log.Factory, cfg *viper.Viper) *Repository {
	return &Repository{
		Repository: pgpool.New(
			pgpool.DefaultDBAlias,
			pgpool.WithName(app),
			pgpool.WithLogger(lg.Zap()),
			pgpool.WithViper(cfg),
		),
		cfg: cfg,
	}
}

func (r *Repository) Messages() repos.MessageRepo {
	if r.messageRepo == nil {
		panic("dev error: Repository not started yet")
	}
	return r.messageRepo
}

func (r *Repository) Start() error {
	if err := r.Repository.Start(); err != nil {
		return err
	}

	// use the postgres implementation of the messageRepo
	r.messageRepo = messages.New(r.DB(), r.Logger(), r.cfg)

	return nil
}
