package nats

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/fredbi/go-experiments/transactional-roundtrip/pkg/injected"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"go.opencensus.io/plugin/ochttp/propagation/tracecontext"
	"go.opencensus.io/trace"
	"go.uber.org/zap"
)

type Server struct {
	rt          injected.Runtime
	ns          *server.Server
	needReloads bool
	reloadDone  chan struct{}
	reloadWg    *sync.WaitGroup
	mx          sync.Mutex

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

	pod, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("could not resolve hostname: %w", err)
	}

	natsOptions := &server.Options{
		ServerName:   fmt.Sprintf("embedded-%s-%s", s.rt.ID(), pod), // must be unique, so we include the pod's hostname
		Host:         host,
		Port:         port,
		NoLog:        !(s.Server.Debug.Logs || s.Server.Debug.Debug || s.Server.Debug.Trace),
		Debug:        s.Server.Debug.Debug,
		Trace:        s.Server.Debug.Trace,
		HTTPPort:     s.Server.MonitorHTTPPort,
		HTTPBasePath: "/probe",
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
			// headless service may not be ready at startup time, so initial route setup should be
			// non-blocking and provided later on, dynamically.
			//
			// we assume that when running from a headless service, all advertised cluster endpoints run on the same port
			lg.Debug("starting headless service discovery", zap.String("headless_service", c.Server.ClusterHeadlessService))
			s.needReloads = true

			natsOptions.Routes, err = discoverIPs(c.Server.ClusterHeadlessService, natsOptions.Cluster.Port, clusterURL.User)
			if err != nil {
				lg.Warn("headless service discovery not available yet", zap.Error(err))
			}
			natsOptions.RoutesStr = ""

		default:
			return errors.New("when running a cluster, you must specify a way to discover the other cluster members: either with explicit routes or a headless service")
		}
		lg.Debug("nats server options", zap.Any("nats_options", natsOptions))

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

	s.ns = ns
	ns.ConfigureLogger()

	go ns.Start()

	if !ns.ReadyForConnections(c.Server.StartupTimeout) {
		return fmt.Errorf("NATS server startup timed out at %s", c.URL)
	}

	if s.needReloads {
		s.startReloader(*natsOptions)
	}

	lg.Info("NATS server started",
		zap.Stringer("url", u),
	)

	return nil
}

func (s *Server) Stop() error {
	if s.ns != nil {
		s.stopReloader()
		s.ns.WaitForShutdown()
	}

	return nil
}

// startReloader hot-reloads the server whenever the headless service changes configuration
func (s *Server) startReloader(natsOptions server.Options) {
	lg := s.rt.Logger().Bg()
	s.reloadDone = make(chan struct{})
	var wg sync.WaitGroup
	s.reloadWg = &wg
	clusterURL, _ := url.Parse(s.Settings.Server.ClusterURL)
	pollInterval := 10 * time.Second

	wg.Add(1)
	go func() {
		defer wg.Done()

		ticker := time.NewTicker(pollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-s.reloadDone:
				return
			case <-ticker.C:
				var err error
				newOptions := natsOptions
				newOptions.Routes, err = discoverIPs(s.Settings.Server.ClusterHeadlessService, newOptions.Cluster.Port, clusterURL.User)
				if err != nil {
					lg.Warn("headless service discovery not available yet", zap.Duration("retry_in", pollInterval), zap.Error(err))

					break
				}
				newOptions.RoutesStr = ""

				if !hasNewRoute(natsOptions.Routes, newOptions.Routes) {
					return
				}

				lg.Debug("reload NATS server with new routes", zap.Stringers("routes", newOptions.Routes))
				s.mx.Lock()
				natsOptions = newOptions
				s.mx.Unlock()

				if err := s.ns.ReloadOptions(&newOptions); err != nil {
					lg.Warn("headless service discovery reloaded wrong server settings", zap.Error(err))
				}
			}
		}
	}()

	lg.Info("started headless service polling", zap.Duration("polling_interval", pollInterval))
}

func hasNewRoute(previous, current []*url.URL) bool {
	prevIdx := make(map[string]struct{}, len(previous))
	for _, route := range previous {
		prevIdx[route.String()] = struct{}{}
	}

	for _, route := range current {
		if _, ok := prevIdx[route.String()]; !ok {
			return true
		}
	}

	return false
}

func (s *Server) stopReloader() {
	if !s.needReloads {
		return
	}

	close(s.reloadDone)

	s.reloadWg.Wait()

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

// discoverIPs performs a DNS reverse-lookup on the headless service to collect the IP addresses of
// the cluster group.
func discoverIPs(svc string, port int, userinfo *url.Userinfo) ([]*url.URL, error) {
	resolver := net.DefaultResolver
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	nsRecord, err := resolver.LookupHost(ctx, svc)
	if err != nil {
		return nil, err
	}

	localIPs, err := getInterfaceIPs()
	if err != nil {
		return nil, err
	}

	nsRecord = filterIPInList(nsRecord, localIPs)

	urls := make([]*url.URL, 0, len(nsRecord))

	for _, addr := range nsRecord {
		u := &url.URL{
			Scheme: "nats",
			Host:   fmt.Sprintf("%v:%d", addr, port),
			User:   userinfo,
		}
		urls = append(urls, u)
	}

	return urls, nil
}

func filterIPInList(unfiltered, filter []string) []string {
	filtered := make([]string, 0, len(unfiltered))
	for _, ip1 := range unfiltered {
		found := false
		for _, ip2 := range filter {
			if ip1 == ip2 {
				found = true

				break
			}
		}
		if found {
			continue
		}

		filtered = append(filtered, ip1)
	}

	return filtered
}

func getInterfaceIPs() ([]string, error) {
	var localIPs []string

	interfaceAddr, err := net.InterfaceAddrs()
	if err != nil {
		return nil, fmt.Errorf("error getting self referencing address: %v", err)
	}

	for i := 0; i < len(interfaceAddr); i++ {
		interfaceIP, _, _ := net.ParseCIDR(interfaceAddr[i].String())
		if net.ParseIP(interfaceIP.String()) != nil {
			localIPs = append(localIPs, interfaceIP.String())
		} else {
			return nil, fmt.Errorf("error parsing self referencing address: %v", err)
		}
	}
	return localIPs, nil
}
