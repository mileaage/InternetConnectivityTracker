package monitor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestWifiLogger(t *testing.T) {
	t.Run("LogConnectivityCheck_Success", func(t *testing.T) {
		tempDir := t.TempDir()
		logFile := filepath.Join(tempDir, "test.log")

		myLogger, err := NewWifiLogger(logFile)

		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}

		deviceID := uuid.NewString()
		now := time.Now()
		err = myLogger.LogConnectivityCheck(deviceID, true, time.Second*2, now, nil)
		if err != nil {
			t.Fatalf("Failed to log connectivity check: %v", err)
		}

		content, err := os.ReadFile(logFile)
		if err != nil {
			t.Fatalf("Failed to read log file: %v", err)
		}

		logString := string(content)
		if !strings.Contains(logString, "CONNECTED IN") {
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
		deviceID := uuid.NewString()
		now := time.Now()
		err = myLogger.LogConnectivityCheck(deviceID, false, time.Second*5, now, testErr)
		if err != nil {
			t.Fatalf("Failed to log connectivity check: %v", err)
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
		deviceID := uuid.NewString()
		start := time.Now().Add(-5 * time.Second)
		end := time.Now()
		err = myLogger.LogOutageStart(deviceID, start)
		if err != nil {
			t.Fatalf("Failed to log outage start: %v", err)
		}
		err = myLogger.LogOutageEnd(deviceID, time.Second*5, end)
		if err != nil {
			t.Fatalf("Failed to log outage end: %v", err)
		}

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
