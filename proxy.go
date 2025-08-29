package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"
)

// proxyServer manages an HTTP reverse proxy for active profile tunnels.
type proxyServer struct {
	ip     string
	routes map[string]int // Host -> local port
	srv    *http.Server
}

var (
	proxyMu sync.Mutex
	proxy   *proxyServer
)

// StartProxy starts an HTTP reverse proxy for the given profile. The proxy
// listens on port 80 of the profile IP address and routes requests based on the
// Host header to the corresponding tunnel's local port.
func StartProxy(p Profile) error {
	proxyMu.Lock()
	defer proxyMu.Unlock()
	if proxy != nil {
		return fmt.Errorf("proxy already running")
	}

	routes := make(map[string]int)
	for _, t := range p.Tunnels {
		routes[t.LocalDomain] = t.LocalPort
	}

	ps := &proxyServer{ip: p.IPAddress, routes: routes}
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", p.IPAddress, 80),
		Handler: http.HandlerFunc(ps.handle),
	}
	ps.srv = srv
	proxy = ps

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("proxy serve: %v", err)
		}
	}()
	return nil
}

// StopProxy stops the running reverse proxy, if any.
func StopProxy() error {
	proxyMu.Lock()
	ps := proxy
	proxy = nil
	proxyMu.Unlock()
	if ps == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return ps.srv.Shutdown(ctx)
}

func (ps *proxyServer) handle(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	if i := strings.Index(host, ":"); i >= 0 {
		host = host[:i]
	}
	port, ok := ps.routes[host]
	if !ok {
		http.Error(w, "no route", http.StatusBadGateway)
		return
	}
	targetURL := &url.URL{Scheme: "http", Host: fmt.Sprintf("%s:%d", ps.ip, port)}
	httputil.NewSingleHostReverseProxy(targetURL).ServeHTTP(w, r)
}
