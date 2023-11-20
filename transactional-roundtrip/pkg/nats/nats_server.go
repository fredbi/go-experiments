package nats

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"

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

	Settings
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

	s.Settings = c
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

	natsOptions := &server.Options{
		ServerName: fmt.Sprintf("embedded-%s", s.rt.ID()),
		Host:       host,
		Port:       port,
		NoLog:      !(s.Server.Debug.Logs || s.Server.Debug.Debug || s.Server.Debug.Trace),
		Debug:      s.Server.Debug.Debug,
		Trace:      s.Server.Debug.Trace,
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

		switch {
		case c.Server.ClusterRoutes != "":
			natsOptions.RoutesStr = c.Server.ClusterRoutes
			routeURLs := server.RoutesFromStr(c.Server.ClusterRoutes)
			if routeURLs == nil {
				return fmt.Errorf("invalid cluster routes: %q", c.Server.ClusterRoutes)
			}
			natsOptions.Routes = routeURLs

		case c.Server.ClusterHeadlessService != "":
			// we assume that when running from a headless service, all advertised cluster endpoints run on the same port
			natsOptions.Routes, err = s.discoverIPs(c.Server.ClusterHeadlessService, natsOptions.Cluster.Port, clusterURL.User)
			if err != nil {
				return err
			}
			natsOptions.RoutesStr = ""

		default:
			return errors.New("when running a cluster, you must specify a way to discover the other cluster members: either with explicit routes or a headless service")
		}

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

func (s Server) discoverIPs(svc string, port int, userinfo *url.Userinfo) ([]*url.URL, error) {
	resolver := net.DefaultResolver
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	nsRecord, err := resolver.LookupIP(ctx, "ip4", svc)
	if err != nil {
		return nil, err
	}

	urls := make([]*url.URL, 0, len(nsRecord))

	for _, addr := range nsRecord {
		u := &url.URL{
			Scheme: "nats",
			Host:   fmt.Sprintf("%v:%d", addr, port),
			User:   userinfo,
		}
		urls = append(urls, u)
	}

	filtered, err := server.RemoveSelfReference(port, urls)
	if err != nil {
		return nil, err
	}

	return filtered, nil
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
