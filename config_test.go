package main

import (
	"os"
	"path/filepath"
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
