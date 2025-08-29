package main

// AuthenticationProvider defines pluggable authentication mechanisms.
type AuthenticationProvider interface {
	Name() string
	Authenticate(t Tunnel) error
}

// EventHandler can react on profile and tunnel events.
type EventHandler interface {
	OnProfileActivated(p Profile)
	OnProfileDeactivated(p Profile)
	OnTunnelFailed(t Tunnel)
}

// Plugin exposes optional capabilities for Lighthouse.
type Plugin interface {
	AuthenticationProvider
	EventHandler
}

var authProvider AuthenticationProvider

// RegisterAuthenticationProvider installs the given authentication provider.
func RegisterAuthenticationProvider(p AuthenticationProvider) {
	authProvider = p
}
