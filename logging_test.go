package main

import (
	"log"
	"os"
	"strings"
	"testing"
)

func TestLogging(t *testing.T) {
	const logPath = "lighthouse.log"
	_ = os.Remove(logPath)
	defer os.Remove(logPath)

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatalf("open log file: %v", err)
	}
	log.SetOutput(logFile)
	log.SetFlags(0)
	testMsg := "test log entry"
	log.Println(testMsg)
	logFile.Close()
	log.SetOutput(os.Stderr)

	if _, err := os.Stat(logPath); err != nil {
		t.Fatalf("log file not found: %v", err)
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}
	if !strings.Contains(string(data), testMsg) {
		t.Fatalf("expected log entry %q not found in %q", testMsg, string(data))
	}
}
