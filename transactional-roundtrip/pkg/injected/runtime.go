package injected

import (
	"github.com/fredbi/go-trace/log"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
)

type Runtime interface {
	Logger() log.Factory
	DB() *sqlx.DB
	Config() *viper.Viper
}
