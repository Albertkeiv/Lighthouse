package main

import (
	"errors"
	"os"
	"os/exec"
)

// SSHAgentPlugin integrates the system SSH agent for authentication.
type SSHAgentPlugin struct{}

// Name returns the plugin name.
func (p *SSHAgentPlugin) Name() string { return "ssh-agent" }

// Authenticate ensures the private key is loaded into the SSH agent.
func (p *SSHAgentPlugin) Authenticate(t Tunnel) error {
	if os.Getenv("SSH_AUTH_SOCK") == "" {
		return errors.New("ssh agent not available")
	}
	if t.SSHKeyPath == "" {
		return errors.New("ssh key path is required")
	}
	cmd := exec.Command("ssh-add", t.SSHKeyPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() {
	RegisterAuthenticationProvider(&SSHAgentPlugin{})
}
