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
	cfg := s.rt.Config()
	lg := s.rt.Logger().Bg()

	c, err := MakeSettings(cfg)
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
		Host: host,
		Port: port,
		Cluster: server.ClusterOpts{
			Name: c.ClusterID,
		},
	}

	ns, err := server.NewServer(natsOptions)
	if err != nil {
		return err
	}

	go ns.Start()

	if !ns.ReadyForConnections(c.Server.StartupTimeout) {
		return fmt.Errorf("NATS server startup timed out at %s", c.URL)
	}

	lg.Info("NATS server started", zap.Stringer("url", u))

	return nil
}

func (s *Server) Stop() error {
	s.ns.WaitForShutdown()

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
