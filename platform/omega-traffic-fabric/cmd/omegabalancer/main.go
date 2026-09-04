package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"pefy.gg/omega-traffic-fabric/internal/balancer"
)

type metrics struct {
	accepted atomic.Uint64
	active   atomic.Int64
	failed   atomic.Uint64
	proxied  atomic.Uint64
}

func main() {
	listen := flag.String("listen", envOr("OMEGA_LISTEN", ":8080"), "TCP listen address")
	rawBackends := flag.String("backends", envOr("OMEGA_BACKENDS", "127.0.0.1:8081,127.0.0.1:8082"), "comma-separated backends")
	algorithm := flag.String("algorithm", envOr("OMEGA_ALGORITHM", "p2c"), "round-robin|least-connections|p2c|maglev")
	metricsAddr := flag.String("metrics", envOr("OMEGA_METRICS", ":9090"), "metrics/health HTTP address; empty disables")
	healthEvery := flag.Duration("health-interval", 5*time.Second, "active TCP health-check interval")
	dialTimeout := flag.Duration("dial-timeout", 2*time.Second, "backend dial timeout")
	drainTimeout := flag.Duration("drain-timeout", 30*time.Second, "graceful drain timeout")
	flag.Parse()

	addrs := splitBackends(*rawBackends)
	pool, err := balancer.NewPool(addrs, balancer.Algorithm(*algorithm))
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go activeHealthChecks(ctx, pool, *healthEvery, *dialTimeout)

	m := &metrics{}
	if *metricsAddr != "" {
		go serveObservability(ctx, *metricsAddr, pool, m)
	}

	ln, err := net.Listen("tcp", *listen)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("ΩBALANCER bootstrap data plane listening on %s with %d backends using %s", *listen, len(pool.Backends()), *algorithm)

	var wg sync.WaitGroup
	go func() {
		<-ctx.Done()
		_ = ln.Close()
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			if ctx.Err() != nil {
				break
			}
			log.Printf("accept error: %v", err)
			continue
		}
		m.accepted.Add(1)
		wg.Add(1)
		go func(c net.Conn) {
			defer wg.Done()
			handleConn(ctx, c, pool, m, *dialTimeout)
		}(conn)
	}

	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(*drainTimeout):
		log.Printf("drain timeout reached; terminating with %d active connections", m.active.Load())
	}
}

func splitBackends(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
		}
	}
	return out
}

func handleConn(ctx context.Context, downstream net.Conn, pool *balancer.Pool, m *metrics, timeout time.Duration) {
	defer downstream.Close()
	backend := pool.Pick(downstream.RemoteAddr().String())
	if backend == nil {
		m.failed.Add(1)
		return
	}
	backend.Acquire()
	defer backend.Release()
	m.active.Add(1)
	defer m.active.Add(-1)

	upstream, err := net.DialTimeout("tcp", backend.Addr, timeout)
	if err != nil {
		backend.SetHealthy(false)
		m.failed.Add(1)
		return
	}
	defer upstream.Close()

	copyDone := make(chan struct{}, 2)
	go copyStream(upstream, downstream, copyDone)
	go copyStream(downstream, upstream, copyDone)
	select {
	case <-copyDone:
	case <-ctx.Done():
	}
	m.proxied.Add(1)
}

func copyStream(dst, src net.Conn, done chan<- struct{}) {
	_, _ = io.Copy(dst, src)
	if tcp, ok := dst.(*net.TCPConn); ok {
		_ = tcp.CloseWrite()
	}
	done <- struct{}{}
}

func activeHealthChecks(ctx context.Context, pool *balancer.Pool, every, timeout time.Duration) {
	ticker := time.NewTicker(every)
	defer ticker.Stop()
	check := func() {
		for _, b := range pool.Backends() {
			c, err := net.DialTimeout("tcp", b.Addr, timeout)
			if err == nil {
				_ = c.Close()
			}
			b.SetHealthy(err == nil)
		}
	}
	check()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			check()
		}
	}
}

func serveObservability(ctx context.Context, addr string, pool *balancer.Pool, m *metrics) {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		healthy := 0
		for _, b := range pool.Backends() {
			if b.Healthy() {
				healthy++
			}
		}
		if healthy == 0 {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		_, _ = fmt.Fprintf(w, "healthy_backends=%d total_backends=%d\n", healthy, len(pool.Backends()))
	})
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "omega_connections_accepted_total %d\n", m.accepted.Load())
		_, _ = fmt.Fprintf(w, "omega_connections_active %d\n", m.active.Load())
		_, _ = fmt.Fprintf(w, "omega_proxy_failures_total %d\n", m.failed.Load())
		_, _ = fmt.Fprintf(w, "omega_connections_proxied_total %d\n", m.proxied.Load())
		for i, b := range pool.Backends() {
			health := 0
			if b.Healthy() {
				health = 1
			}
			_, _ = fmt.Fprintf(w, "omega_backend_health{backend=\"%d\"} %d\n", i, health)
			_, _ = fmt.Fprintf(w, "omega_backend_active_connections{backend=\"%d\"} %d\n", i, b.Active())
		}
	})
	srv := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 3 * time.Second}
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdown)
	}()
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("observability server: %v", err)
	}
}

func envOr(k, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return fallback
}
