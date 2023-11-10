package keys

import (
	"github.com/spf13/viper"
)

const (
	URL           = "url"
	ClusterID     = "cluster_id"
	PostingsTopic = "topics.postings"
	ResultsTopic  = "topics.results"
	Port          = "port"
)

func SetDefaults(cfg *viper.Viper) {
	// TODO
}
