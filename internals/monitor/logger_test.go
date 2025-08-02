package monitor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWifiLogger(t *testing.T) {
	t.Run("LogConnectivityCheck_Success", func(t *testing.T) {
		tempDir := t.TempDir()
		logFile := filepath.Join(tempDir, "test.log")

		myLogger, err := NewWifiLogger(logFile)

		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}

		myLogger.LogConnectivityCheck(true, time.Second*2, nil)
		if err != nil {
			t.Fatalf("Failed to log: %v", err)
		}

		content, err := os.ReadFile(logFile)
		if err != nil {
			t.Fatalf("Failed to read log file: %v", err)
		}

		logString := string(content)
		if !strings.Contains(logString, "CONNECTED TO google.com") {
			t.Errorf("Expected success log, got: %s", logString)
		}
	})

	t.Run("LogConnectivityCheck_Failure", func(t *testing.T) {
		tempDir := t.TempDir()
		logFile := filepath.Join(tempDir, "test.log")

		myLogger, err := NewWifiLogger(logFile)
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}

		testErr := fmt.Errorf("connection timeout")
		myLogger.LogConnectivityCheck(false, time.Second*5, testErr)
		if err != nil {
			t.Fatalf("Failed to log: %v", err)
		}

		content, err := os.ReadFile(logFile)
		if err != nil {
			t.Fatalf("Failed to read log file: %v", err)
		}

		logString := string(content)
		if !strings.Contains(logString, "FAILED TO CONNECT") {
			t.Errorf("Expected failure log, got: %s", logString)
		}
	})

	t.Run("LogOutageCheck_Success", func(t *testing.T) {
		tempDir := t.TempDir()
		logFile := filepath.Join(tempDir, "test.log")

		myLogger, err := NewWifiLogger(logFile)
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}

		// first we log the outage
		myLogger.LogOutageStart(time.Now().Add(-5 * time.Second))

		myLogger.LogOutageEnd(time.Second*5, time.Now())

		content, err := os.ReadFile(logFile)
		if err != nil {
			t.Fatalf("Failed to read log file: %v", err)
		}

		logString := string(content)
		if !strings.Contains(logString, "OUTAGE END") {
			t.Errorf("Expected outage log end, got: %s", logString)
		}
	})

}
