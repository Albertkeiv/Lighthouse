package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

const defaultConfigFile = "profiles.json"

// configDir returns the directory where user specific configuration is stored.
// The location varies depending on the operating system:
//
//	Windows:   directory where the program is executed
//	Linux/macOS/Android: ~/.config/lighthouse
//	others:    current working directory
//
// If the home directory cannot be determined, the current working directory is
// used as a fallback so that profiles can still be read and written.
func configDir() (string, error) {
	switch runtime.GOOS {
	case "windows":
		return os.Getwd()
	case "linux", "darwin", "android":
		home, err := os.UserHomeDir()
		if err != nil {
			wd, _ := os.Getwd()
			return wd, nil
		}
		return filepath.Join(home, ".config", "lighthouse"), nil
	default:
		return os.Getwd()
	}
}

// LoadProfiles reads profiles from the configuration file.
func LoadProfiles() ([]Profile, error) {
	dir, err := configDir()
	if err != nil {
		log.Printf("configDir: %v", err)
		return nil, err
	}
	path := filepath.Join(dir, defaultConfigFile)
	log.Printf("loading profiles from %s", path)
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		log.Printf("config file %s not found", path)
		return []Profile{}, nil
	}
	if err != nil {
		log.Printf("read config file %s: %v", path, err)
		return nil, err
	}
	var profiles []Profile
	if err := json.Unmarshal(data, &profiles); err != nil {
		log.Printf("parse config file %s: %v", path, err)
		return nil, err
	}
	for _, p := range profiles {
		if err := p.Validate(); err != nil {
			log.Printf("invalid profile %s: %v", p.Name, err)
			return nil, fmt.Errorf("invalid profile %s: %w", p.Name, err)
		}
		log.Printf("loaded profile %s", p.Name)
	}
	return profiles, nil
}

// SaveProfiles writes profiles to the configuration file.
func SaveProfiles(profiles []Profile) error {
	dir, err := configDir()
	if err != nil {
		return err
	}
	path := filepath.Join(dir, defaultConfigFile)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			log.Printf("create config directory %s: %v", dir, err)
			return err
		}
		log.Printf("created config directory %s", dir)
	}

	for _, p := range profiles {
		if err := p.Validate(); err != nil {
			return fmt.Errorf("invalid profile %s: %w", p.Name, err)
		}
		log.Printf("saving profile %s", p.Name)
	}

	data, err := json.MarshalIndent(profiles, "", "  ")
	if err != nil {
		return err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Printf("creating config file %s", path)
	} else if err == nil {
		log.Printf("updating config file %s", path)
	} else {
		log.Printf("stat config file %s: %v", path, err)
		return err
	}

	return os.WriteFile(path, data, 0o600)
}
