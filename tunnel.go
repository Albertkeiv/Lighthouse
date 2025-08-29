package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

type tunnelState struct {
	cancel context.CancelFunc
	done   chan struct{}
}

var (
	tunnelStates = make(map[string]*tunnelState)
	tunnelMu     sync.Mutex
)

// StartTunnel launches the SSH tunnel described by t. It listens on the
// configured local address and forwards connections to the remote host. The
// tunnel automatically tries to reconnect if the SSH connection drops.
func StartTunnel(t Tunnel) error {
	tunnelMu.Lock()
	if _, ok := tunnelStates[t.Name]; ok {
		tunnelMu.Unlock()
		return fmt.Errorf("tunnel %s already running", t.Name)
	}
	ctx, cancel := context.WithCancel(context.Background())
	ts := &tunnelState{cancel: cancel, done: make(chan struct{})}
	tunnelStates[t.Name] = ts
	tunnelMu.Unlock()
	log.Printf("starting tunnel %s", t.Name)
	go func() {
		defer close(ts.done)
		for {
			if err := runTunnel(ctx, t); err != nil {
				log.Printf("tunnel %s error: %v", t.Name, err)
			}
			if ctx.Err() != nil {
				return
			}
			log.Printf("tunnel %s restarting", t.Name)
			time.Sleep(time.Second)
		}
	}()
	return nil
}

// StopTunnel stops the running tunnel described by t.
func StopTunnel(t Tunnel) error {
	tunnelMu.Lock()
	ts, ok := tunnelStates[t.Name]
	if ok {
		delete(tunnelStates, t.Name)
	}
	tunnelMu.Unlock()
	if !ok {
		return nil
	}
	ts.cancel()
	<-ts.done
	log.Printf("stopped tunnel %s", t.Name)
	return nil
}

func runTunnel(ctx context.Context, t Tunnel) error {
	localAddr := fmt.Sprintf("%s:%d", t.LocalDomain, t.LocalPort)
	listener, err := net.Listen("tcp", localAddr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	defer listener.Close()

	for {
		if ctx.Err() != nil {
			return nil
		}
		client, err := dialSSH(t)
		if err != nil {
			log.Printf("tunnel %s dial: %v", t.Name, err)
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(time.Second):
			}
			continue
		}
		err = serveTunnel(ctx, listener, client, t)
		client.Close()
		if err != nil {
			log.Printf("tunnel %s serve: %v", t.Name, err)
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Second):
		}
	}
}

func dialSSH(t Tunnel) (*ssh.Client, error) {
	key, err := os.ReadFile(t.SSHKeyPath)
	if err != nil {
		return nil, err
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}
	config := &ssh.ClientConfig{
		User:            t.SSHUser,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}
	addr := fmt.Sprintf("%s:%d", t.SSHServer, t.SSHPort)
	return ssh.Dial("tcp", addr, config)
}

func serveTunnel(ctx context.Context, l net.Listener, client *ssh.Client, t Tunnel) error {
	connCh := make(chan net.Conn)
	errCh := make(chan error, 1)

	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				errCh <- err
				return
			}
			connCh <- c
		}
	}()

	go func() {
		errCh <- client.Conn.Wait()
	}()

	remoteAddr := fmt.Sprintf("%s:%d", t.RemoteHost, t.RemotePort)
	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-errCh:
			return err
		case lc := <-connCh:
			go handleConn(lc, client, remoteAddr)
		}
	}
}

func handleConn(lc net.Conn, client *ssh.Client, remoteAddr string) {
	rc, err := client.Dial("tcp", remoteAddr)
	if err != nil {
		lc.Close()
		return
	}
	go func() {
		io.Copy(lc, rc)
		lc.Close()
		rc.Close()
	}()
	go func() {
		io.Copy(rc, lc)
		lc.Close()
		rc.Close()
	}()
}

// IsTunnelRunning reports whether the specified tunnel is currently active.
func IsTunnelRunning(t Tunnel) bool {
	tunnelMu.Lock()
	defer tunnelMu.Unlock()
	_, ok := tunnelStates[t.Name]
	return ok
}
