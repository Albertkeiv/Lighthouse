package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// TestConfigDirFallback verifies that profiles can be saved and loaded when
// the user configuration directory cannot be determined.
func TestConfigDirFallback(t *testing.T) {
	// Unset environment variables so os.UserConfigDir fails.
	origHome := os.Getenv("HOME")
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Unsetenv("HOME")
	os.Unsetenv("XDG_CONFIG_HOME")
	defer os.Setenv("HOME", origHome)
	if origXDG != "" {
		defer os.Setenv("XDG_CONFIG_HOME", origXDG)
	}

	// Use a temporary working directory.
	dir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer os.Chdir(oldWd)
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	// Saving and loading an empty profile list should succeed and use the
	// working directory for storage.
	if err := SaveProfiles([]Profile{}); err != nil {
		t.Fatalf("SaveProfiles: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, defaultConfigFile)); err != nil {
		t.Fatalf("expected config file in working directory: %v", err)
	}
	profiles, err := LoadProfiles()
	if err != nil {
		t.Fatalf("LoadProfiles: %v", err)
	}
	if len(profiles) != 0 {
		t.Fatalf("expected zero profiles, got %d", len(profiles))
	}
}

// TestProfileTunnelPersistence ensures that tunnels are preserved across save and load operations.
func TestProfileTunnelPersistence(t *testing.T) {
	// Unset environment variables so os.UserConfigDir fails and configDir uses the working directory.
	origHome := os.Getenv("HOME")
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Unsetenv("HOME")
	os.Unsetenv("XDG_CONFIG_HOME")
	defer os.Setenv("HOME", origHome)
	if origXDG != "" {
		defer os.Setenv("XDG_CONFIG_HOME", origXDG)
	}

	// Use a temporary working directory.
	dir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer os.Chdir(oldWd)
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	original := Profile{
		Name:      "p1",
		IPAddress: "127.0.0.1",
		Tunnels: []Tunnel{
			{
				Name:        "t1",
				SSHServer:   "ssh.example.com",
				SSHPort:     22,
				SSHUser:     "user1",
				SSHKeyPath:  "/path/to/key1",
				RemoteHost:  "remote1",
				RemotePort:  80,
				LocalDomain: "local1",
				LocalPort:   8080,
			},
			{
				Name:        "t2",
				SSHServer:   "ssh.example.org",
				SSHPort:     2222,
				SSHUser:     "user2",
				SSHKeyPath:  "/path/to/key2",
				RemoteHost:  "remote2",
				RemotePort:  443,
				LocalDomain: "local2",
				LocalPort:   8443,
			},
		},
	}

	if err := SaveProfiles([]Profile{original}); err != nil {
		t.Fatalf("SaveProfiles: %v", err)
	}

	loaded, err := LoadProfiles()
	if err != nil {
		t.Fatalf("LoadProfiles: %v", err)
	}

	if len(loaded) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(loaded))
	}
	if !reflect.DeepEqual(original, loaded[0]) {
		t.Fatalf("loaded profile does not match original\noriginal: %#v\nloaded: %#v", original, loaded[0])
	}
}

// TestAddTunnelSavesProfile verifies that adding a tunnel to a profile also
// persists the updated profile to disk.
func TestAddTunnelSavesProfile(t *testing.T) {
	// Unset environment variables so os.UserConfigDir fails and configDir uses the working directory.
	origHome := os.Getenv("HOME")
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Unsetenv("HOME")
	os.Unsetenv("XDG_CONFIG_HOME")
	defer os.Setenv("HOME", origHome)
	if origXDG != "" {
		defer os.Setenv("XDG_CONFIG_HOME", origXDG)
	}

	// Use a temporary working directory.
	dir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer os.Chdir(oldWd)
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	p := Profile{Name: "p1", IPAddress: "127.0.0.1"}
	if err := SaveProfiles([]Profile{p}); err != nil {
		t.Fatalf("SaveProfiles: %v", err)
	}

	tnl := Tunnel{
		Name:        "t1",
		SSHServer:   "ssh.example.com",
		SSHPort:     22,
		SSHUser:     "user",
		SSHKeyPath:  "/path/to/key",
		RemoteHost:  "remote",
		RemotePort:  80,
		LocalDomain: "local",
		LocalPort:   8080,
	}

	if err := p.AddTunnel(tnl); err != nil {
		t.Fatalf("AddTunnel: %v", err)
	}

	loaded, err := LoadProfiles()
	if err != nil {
		t.Fatalf("LoadProfiles: %v", err)
	}
	if len(loaded) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(loaded))
	}
	if len(loaded[0].Tunnels) != 1 {
		t.Fatalf("expected 1 tunnel, got %d", len(loaded[0].Tunnels))
	}
	if !reflect.DeepEqual(tnl, loaded[0].Tunnels[0]) {
		t.Fatalf("loaded tunnel does not match added tunnel\nexpected: %#v\nloaded: %#v", tnl, loaded[0].Tunnels[0])
	}
}
