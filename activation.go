package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"sync"
)

var (
	activeProfile *Profile
	activeCmds    []*exec.Cmd
	mu            sync.Mutex
)

// ActivateProfile starts all tunnels defined in the given profile. Any previously
// active profile is deactivated to ensure that only a single profile is running
// at a time.
func ActivateProfile(p Profile) error {
	mu.Lock()
	defer mu.Unlock()

	// If another profile is active, deactivate it first.
	if activeProfile != nil && activeProfile.Name != p.Name {
		if err := deactivateLocked(); err != nil {
			return err
		}
	} else if activeProfile != nil && activeProfile.Name == p.Name {
		// Already active; nothing to do.
		return nil
	}

	var cmds []*exec.Cmd
	for _, t := range p.Tunnels {
		if authProvider != nil {
			if err := authProvider.Authenticate(t); err != nil {
				for _, c := range cmds {
					if c.Process != nil {
						_ = c.Process.Kill()
					}
				}
				return err
			}
		}

		args := []string{
			"-o", "PreferredAuthentications=publickey",
			"-o", "PasswordAuthentication=no",
			"-N",
			"-L", fmt.Sprintf("%s:%d:%s:%d", t.LocalDomain, t.LocalPort, t.RemoteHost, t.RemotePort),
		}
		if t.SSHKeyPath != "" {
			args = append(args, "-i", t.SSHKeyPath)
		}
		args = append(args, "-p", strconv.Itoa(t.SSHPort), fmt.Sprintf("%s@%s", t.SSHUser, t.SSHServer))
		cmd := exec.Command("ssh", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Start(); err != nil {
			// stop any started tunnels on error
			for _, c := range cmds {
				if c.Process != nil {
					_ = c.Process.Kill()
				}
			}
			return err
		}
		cmds = append(cmds, cmd)
	}

	activeProfile = &p
	activeCmds = cmds
	return nil
}

// DeactivateProfile stops all tunnels of the currently active profile, if any.
func DeactivateProfile() error {
	mu.Lock()
	defer mu.Unlock()
	return deactivateLocked()
}

// deactivateLocked assumes the caller holds mu.
func deactivateLocked() error {
	if activeProfile == nil {
		return nil
	}
	for _, c := range activeCmds {
		if c.Process != nil {
			if err := c.Process.Kill(); err != nil {
				return err
			}
			_, _ = c.Process.Wait()
		}
	}
	activeProfile = nil
	activeCmds = nil
	return nil
}
