package main

import (
	"errors"
	"fmt"
	"net"
)

// Tunnel represents a single SSH tunnel configuration within a profile.
type Tunnel struct {
	Name        string `json:"name" yaml:"name"`
	SSHServer   string `json:"ssh_server" yaml:"ssh_server"`
	SSHPort     int    `json:"ssh_port" yaml:"ssh_port"`
	SSHUser     string `json:"ssh_user" yaml:"ssh_user"`
	SSHKeyPath  string `json:"ssh_key_path" yaml:"ssh_key_path"`
	RemoteHost  string `json:"remote_host" yaml:"remote_host"`
	RemotePort  int    `json:"remote_port" yaml:"remote_port"`
	LocalDomain string `json:"local_domain" yaml:"local_domain"`
	LocalPort   int    `json:"local_port" yaml:"local_port"`
}

// Validate checks that all required tunnel fields are present.
func (t Tunnel) Validate() error {
	if t.Name == "" {
		return errors.New("tunnel name is required")
	}
	if t.SSHServer == "" {
		return errors.New("ssh_server is required")
	}
	if t.SSHPort <= 0 {
		return errors.New("ssh_port must be > 0")
	}
	if t.SSHUser == "" {
		return errors.New("ssh_user is required")
	}
	if t.SSHKeyPath == "" {
		return errors.New("ssh_key_path is required")
	}
	if t.RemoteHost == "" {
		return errors.New("remote_host is required")
	}
	if t.RemotePort <= 0 {
		return errors.New("remote_port must be > 0")
	}
	if t.LocalDomain == "" {
		return errors.New("local_domain is required")
	}
	if t.LocalPort <= 0 {
		return errors.New("local_port must be > 0")
	}
	return nil
}

// Profile groups multiple tunnels under a single IP address.
type Profile struct {
	Name      string   `json:"name" yaml:"name"`
	IPAddress string   `json:"ip_address" yaml:"ip_address"`
	Tunnels   []Tunnel `json:"tunnels" yaml:"tunnels"`
}

// Validate ensures the profile has a loopback IP and valid tunnels.
func (p Profile) Validate() error {
	ip := net.ParseIP(p.IPAddress)
	if ip == nil {
		return fmt.Errorf("invalid ip address: %s", p.IPAddress)
	}
	ip4 := ip.To4()
	if ip4 == nil || ip4[0] != 127 {
		return fmt.Errorf("ip address must be in 127.0.0.0/8: %s", p.IPAddress)
	}
	for _, t := range p.Tunnels {
		if err := t.Validate(); err != nil {
			return fmt.Errorf("tunnel %s: %w", t.Name, err)
		}
	}
	return nil
}

// AddTunnel appends a new tunnel to the profile and persists the updated
// profile to disk. The tunnel configuration is validated before saving.
func (p *Profile) AddTunnel(t Tunnel) error {
	if err := t.Validate(); err != nil {
		return fmt.Errorf("invalid tunnel %s: %w", t.Name, err)
	}
	p.Tunnels = append(p.Tunnels, t)
	return SaveProfile(*p)
}
