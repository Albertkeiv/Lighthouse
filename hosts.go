package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
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

// AddHostEntry appends a mapping of ip to domain in the hosts file.
// The entry will look like: "<IP> <domain>".
// If the entry already exists, the function returns without error.
// If writing requires elevated privileges, sudo will be invoked automatically.
func AddHostEntry(ip, domain string) error {
	path := hostsFile()
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
		if os.IsPermission(err) {
			cmd := exec.Command("sudo", "tee", "-a", path)
			cmd.Stdin = strings.NewReader(entry + "\n")
			return cmd.Run()
		}
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
// If writing requires elevated privileges, sudo will be invoked automatically.
func RemoveHostEntries(ip, domain string) error {
	path := hostsFile()
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

	if err := writeHostsFile(path, buf.Bytes()); err != nil {
		return err
	}
	return nil
}

// writeHostsFile writes content to the hosts file handling sudo when needed.
func writeHostsFile(path string, content []byte) error {
	if err := os.WriteFile(path, content, 0o644); err != nil {
		if os.IsPermission(err) {
			cmd := exec.Command("sudo", "tee", path)
			cmd.Stdin = bytes.NewReader(content)
			return cmd.Run()
		}
		return err
	}
	return nil
}
