package nats

import (
	"fmt"
	"net"
	"net/url"
	"strconv"

	configkeys "github.com/fredbi/go-experiments/transactional-roundtrip/cmd/daemon/commands/config-keys"
	"github.com/fredbi/go-experiments/transactional-roundtrip/pkg/injected"
	natsconfigkeys "github.com/fredbi/go-experiments/transactional-roundtrip/pkg/nats/config-keys"
	"github.com/nats-io/nats-server/v2/server"
	"go.uber.org/zap"
)

type Server struct {
	rt injected.Runtime
	ns *server.Server
}

func New(rt injected.Runtime) *Server {
	return &Server{
		rt: rt,
	}
}

// Start an embedded NATS server in the background.
func (s *Server) Start() error {
	cfg := s.rt.Config()
	lg := s.rt.Logger().Bg()

	natsConfig := cfg.Sub(configkeys.NatsConfig)
	if natsConfig == nil {
		natsConfig = natsconfigkeys.DefaultNATSConfig()
	}

	natsURL := natsConfig.GetString(natsconfigkeys.URL)
	clusterID := natsConfig.GetString(natsconfigkeys.ClusterID)
	startupTimeout := natsConfig.GetDuration(natsconfigkeys.StartupTimeout)

	u, err := url.Parse(natsURL)
	if err != nil {
		return err
	}

	host, portStr, err := net.SplitHostPort(u.Host)
	if err != nil {
		return err
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return err
	}

	// TODO: investigate the impact of various options
	natsOptions := &server.Options{
		Host: host,
		Port: port,
		Cluster: server.ClusterOpts{
			Name: clusterID,
		},
	}

	ns, err := server.NewServer(natsOptions)
	if err != nil {
		return err
	}

	go ns.Start()

	if !ns.ReadyForConnections(startupTimeout) {
		return fmt.Errorf("NATS server startup timed out at %s", natsURL)
	}

	lg.Info("NATS server started", zap.Stringer("url", u))

	return nil
}

func (s *Server) Stop() error {
	s.ns.WaitForShutdown()

	return nil
}
