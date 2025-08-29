package main

import (
	"sync"
)

var (
	activeProfile *Profile
	activeTunnels []Tunnel
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

	var started []Tunnel
	for _, t := range p.Tunnels {
		if authProvider != nil {
			if err := authProvider.Authenticate(t); err != nil {
				for _, st := range started {
					_ = StopTunnel(st)
				}
				return err
			}
		}

		if err := StartTunnel(t); err != nil {
			for _, st := range started {
				_ = StopTunnel(st)
			}
			return err
		}
		started = append(started, t)
	}

	activeProfile = &p
	activeTunnels = started
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
	for _, t := range activeTunnels {
		if err := StopTunnel(t); err != nil {
			return err
		}
	}
	activeProfile = nil
	activeTunnels = nil
	return nil
}
