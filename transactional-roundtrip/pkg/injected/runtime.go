package injected

import (
	"github.com/fredbi/go-experiments/transactional-roundtrip/pkg/repos"
	"github.com/fredbi/go-trace/log"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
)

type Runtime interface {
	Logger() log.Factory
	DB() *sqlx.DB
	Config() *viper.Viper
	Repos() repos.Iface
}
