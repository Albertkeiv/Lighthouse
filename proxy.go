package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

type proxyListener struct {
	net.Listener
	localPort int
}

type proxyServer struct {
	ip        string
	listeners []proxyListener
	ctx       context.Context
	cancel    context.CancelFunc
}

var (
	proxyMu sync.Mutex
	proxy   *proxyServer
)

// StartProxy starts TCP listeners for each tunnel, forwarding connections to
// the corresponding local ports.
func StartProxy(p Profile) error {
	proxyMu.Lock()
	defer proxyMu.Unlock()
	if proxy != nil {
		return fmt.Errorf("proxy already running")
	}

	ctx, cancel := context.WithCancel(context.Background())
	ps := &proxyServer{ip: p.IPAddress, ctx: ctx, cancel: cancel}

	for _, t := range p.Tunnels {
		addr := fmt.Sprintf("%s:%d", p.IPAddress, t.RemotePort)
		l, err := net.Listen("tcp", addr)
		if err != nil {
			cancel()
			for _, pl := range ps.listeners {
				pl.Close()
			}
			return err
		}
		pl := proxyListener{Listener: l, localPort: t.LocalPort}
		ps.listeners = append(ps.listeners, pl)
		go ps.acceptLoop(ctx, pl)
	}

	proxy = ps
	return nil
}

// StopProxy stops all running TCP listeners.
func StopProxy() error {
	proxyMu.Lock()
	ps := proxy
	proxy = nil
	proxyMu.Unlock()
	if ps == nil {
		return nil
	}

	ps.cancel()
	var firstErr error
	for _, pl := range ps.listeners {
		if err := pl.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (ps *proxyServer) acceptLoop(ctx context.Context, pl proxyListener) {
	for {
		c, err := pl.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				log.Printf("proxy accept: %v", err)
				continue
			}
		}
		go ps.handleConn(ctx, c, pl.localPort)
	}
}

func (ps *proxyServer) handleConn(ctx context.Context, conn net.Conn, port int) {
	dstAddr := fmt.Sprintf("%s:%d", ps.ip, port)
	dst, err := net.Dial("tcp", dstAddr)
	if err != nil {
		log.Printf("proxy dial: %v", err)
		conn.Close()
		return
	}

	ctx, cancel := context.WithCancel(ctx)
	go func() {
		<-ctx.Done()
		conn.Close()
		dst.Close()
	}()
	go func() {
		io.Copy(dst, conn)
		cancel()
	}()
	go func() {
		io.Copy(conn, dst)
		cancel()
	}()
	<-ctx.Done()
}
