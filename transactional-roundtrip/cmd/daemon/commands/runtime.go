package commands

import (
	"github.com/fredbi/go-experiments/transactional-roundtrip/pkg/injected"
	"github.com/fredbi/go-trace/log"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
)

type runtime struct {
	db     *sqlx.DB
	logger log.Factory
	cfg    *viper.Viper
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
