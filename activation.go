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

	var (
		started []Tunnel
		added   []Tunnel
	)
	for _, t := range p.Tunnels {
		if err := AddHostEntry(p.IPAddress, t.LocalDomain); err != nil {
			for _, at := range added {
				_ = RemoveHostEntries(p.IPAddress, at.LocalDomain)
			}
			return err
		}
		added = append(added, t)

		if authProvider != nil {
			if err := authProvider.Authenticate(t); err != nil {
				for _, st := range started {
					_ = StopTunnel(st)
				}
				for _, at := range added {
					_ = RemoveHostEntries(p.IPAddress, at.LocalDomain)
				}
				return err
			}
		}

		if err := StartTunnel(t); err != nil {
			for _, st := range started {
				_ = StopTunnel(st)
			}
			for _, at := range added {
				_ = RemoveHostEntries(p.IPAddress, at.LocalDomain)
			}
			return err
		}
		started = append(started, t)
	}

	if err := StartProxy(p); err != nil {
		for _, st := range started {
			_ = StopTunnel(st)
		}
		for _, at := range added {
			_ = RemoveHostEntries(p.IPAddress, at.LocalDomain)
		}
		return err
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
		_ = RemoveHostEntries(activeProfile.IPAddress, t.LocalDomain)
		if err := StopTunnel(t); err != nil {
			return err
		}
	}
	if err := StopProxy(); err != nil {
		return err
	}
	activeProfile = nil
	activeTunnels = nil
	return nil
}

// IsProfileActive returns true if the given profile is currently active.
func IsProfileActive(p Profile) bool {
	mu.Lock()
	defer mu.Unlock()
	return activeProfile != nil && activeProfile.Name == p.Name
}
