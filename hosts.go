package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// hostsFile returns the path to the system hosts file depending on the OS.
func hostsFile() string {
	if runtime.GOOS == "windows" {
		winDir := os.Getenv("WINDIR")
		if winDir == "" {
			winDir = `C:\\Windows`
		}
		return filepath.Join(winDir, "System32", "drivers", "etc", "hosts")
	}
	return "/etc/hosts"
}

// validateHostsPermissions checks if the hosts file is writable.
func validateHostsPermissions(path string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0)
	if err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("insufficient permissions to modify hosts file: %w", err)
		}
		return err
	}
	f.Close()
	return nil
}

// AddHostEntry appends a mapping of ip to domain in the hosts file.
// The entry will look like: "<IP> <domain>".
// If the entry already exists, the function returns without error.
// Returns an error if the hosts file is not writable.
func AddHostEntry(ip, domain string) error {
	path := hostsFile()
	if err := validateHostsPermissions(path); err != nil {
		return err
	}
	entry := fmt.Sprintf("%s %s", ip, domain)

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if strings.Contains(string(data), entry) {
		return nil
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := fmt.Fprintln(f, entry); err != nil {
		return err
	}
	return nil
}

// RemoveHostEntries removes all host file entries that match the provided ip.
// If domain is not empty, only entries with both ip and domain will be removed.
// Returns an error if the hosts file is not writable.
func RemoveHostEntries(ip, domain string) error {
	path := hostsFile()
	if err := validateHostsPermissions(path); err != nil {
		return err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	var buf bytes.Buffer
	removed := false
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) > 0 && fields[0] == ip {
			if domain == "" || (len(fields) > 1 && fields[1] == domain) {
				removed = true
				continue
			}
		}
		buf.WriteString(line)
		buf.WriteByte('\n')
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	if !removed {
		return nil
	}

	return writeHostsFile(path, buf.Bytes())
}

// writeHostsFile writes content to the hosts file.
func writeHostsFile(path string, content []byte) error {
	if err := os.WriteFile(path, content, 0o644); err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("insufficient permissions to modify hosts file: %w", err)
		}
		return err
	}
	return nil
}
