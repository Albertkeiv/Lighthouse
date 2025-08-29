package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

const defaultConfigFile = "profiles.json"

// configDir returns the directory where user specific configuration is stored.
func configDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "lighthouse"), nil
}

// LoadProfiles reads profiles from the configuration file.
func LoadProfiles() ([]Profile, error) {
	dir, err := configDir()
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadFile(filepath.Join(dir, defaultConfigFile))
	if errors.Is(err, os.ErrNotExist) {
		return []Profile{}, nil
	}
	if err != nil {
		return nil, err
	}
	var profiles []Profile
	if err := json.Unmarshal(data, &profiles); err != nil {
		return nil, err
	}
	for _, p := range profiles {
		if err := p.Validate(); err != nil {
			return nil, fmt.Errorf("invalid profile %s: %w", p.Name, err)
		}
	}
	return profiles, nil
}

// SaveProfiles writes profiles to the configuration file.
func SaveProfiles(profiles []Profile) error {
	dir, err := configDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	for _, p := range profiles {
		if err := p.Validate(); err != nil {
			return fmt.Errorf("invalid profile %s: %w", p.Name, err)
		}
	}
	data, err := json.MarshalIndent(profiles, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filepath.Join(dir, defaultConfigFile), data, 0o600)
}
