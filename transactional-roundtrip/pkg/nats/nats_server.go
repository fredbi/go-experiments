package nats

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"

	"github.com/fredbi/go-experiments/transactional-roundtrip/pkg/injected"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"go.opencensus.io/plugin/ochttp/propagation/tracecontext"
	"go.opencensus.io/trace"
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
	lg := s.rt.Logger().Bg()

	c, err := MakeSettings(s.rt.Config())
	if err != nil {
		return err
	}

	u, err := url.Parse(c.URL)
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
		ServerName: fmt.Sprintf("embedded-%s", s.rt.ID()),
		Host:       host,
		Port:       port,
		Debug:      true,
	}

	if c.Server.ClusterID != "" {
		clusterURL, e := url.Parse(c.Server.ClusterURL)
		if e != nil {
			return e
		}

		clusterHost, clusterPortStr, e := net.SplitHostPort(clusterURL.Host)
		if e != nil {
			return e
		}

		clusterPort, e := strconv.Atoi(clusterPortStr)
		if e != nil {
			return e
		}

		natsOptions.Cluster = server.ClusterOpts{
			Name:           c.Server.ClusterID,
			Port:           clusterPort,
			Host:           clusterHost,
			ListenStr:      c.Server.ClusterURL,
			ConnectRetries: 100,
		}

		if clusterURL.User != nil {
			natsOptions.Cluster.Username = clusterURL.User.Username()
			if passwd, hasPassword := clusterURL.User.Password(); hasPassword {
				natsOptions.Cluster.Password = passwd
			}
		}

		natsOptions.RoutesStr = c.Server.ClusterRoutes
		routeURLs := server.RoutesFromStr(c.Server.ClusterRoutes)
		if routeURLs == nil {
			return fmt.Errorf("invalid cluster routes: %q", c.Server.ClusterRoutes)
		}
		natsOptions.Routes = routeURLs

		lg.Info("NATS clustering enabled",
			zap.String("cluster_id", natsOptions.Cluster.Name),
			zap.String("cluster_url", natsOptions.Cluster.ListenStr),
			zap.Stringers("cluster_routes", natsOptions.Routes),
		)
	}

	ns, err := server.NewServer(natsOptions)
	if err != nil {
		return err
	}
	ns.ConfigureLogger()

	go ns.Start()

	if !ns.ReadyForConnections(c.Server.StartupTimeout) {
		return fmt.Errorf("NATS server startup timed out at %s", c.URL)
	}

	lg.Info("NATS server started",
		zap.Stringer("url", u),
	)

	return nil
}

func (s *Server) Stop() error {
	if s.ns != nil {
		s.ns.WaitForShutdown()
	}

	return nil
}

var p = &tracecontext.HTTPFormat{}

// SpanContextFromHeaders extracts a trace span from the headers of a NATS message.
func SpanContextFromHeaders(parentCtx context.Context, msg *nats.Msg) context.Context {
	traceID := msg.Header.Get("trace_id")
	spanID := msg.Header.Get("span_id")
	spanCtx, ok := p.SpanContextFromHeaders(traceID, spanID)
	if !ok {
		return parentCtx
	}

	ctx, _ := trace.StartSpanWithRemoteParent(parentCtx, "incoming NATS message", spanCtx)

	return ctx
}
