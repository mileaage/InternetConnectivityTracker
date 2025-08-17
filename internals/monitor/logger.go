// In monitor/logger.go
package monitor

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Wifi Logger
type WifiLogger struct {
	logFile string
}

func NewWifiLogger(logFile string) (*WifiLogger, error) {
	// create the logfile
	dir := filepath.Dir(logFile)

	err := os.MkdirAll(dir, 0755)
	if err != nil {
		panic(err)
	}

	return &WifiLogger{
		logFile: logFile,
	}, nil
}

func (f *WifiLogger) openLogFile() (*os.File, error) {
	openFile, openErr := os.OpenFile(f.logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if openErr != nil {
		return nil, openErr
	}

	return openFile, nil
}

func (f *WifiLogger) LogConnectivityCheck(deviceID string, success bool, responseTime time.Duration, timestamp time.Time, err error) error {
	file, openErr := f.openLogFile()
	if openErr != nil {
		return openErr
	}
	defer file.Close()

	var logMessage string
	if success {
		logMessage = fmt.Sprintf("[%s] DEVICE: %s CONNECTED IN %.2f SECONDS\n", timestamp.Format(time.RFC3339), deviceID, responseTime.Seconds())
	} else {
		logMessage = fmt.Sprintf("[%s] DEVICE: %s FAILED TO CONNECT IN %.2f SECONDS, ERROR: %v\n", timestamp.Format(time.RFC3339), deviceID, responseTime.Seconds(), err)
	}

	_, writeErr := file.WriteString(logMessage)
	return writeErr
}

func (f *WifiLogger) LogStatusChange(deviceID string, from, to ConnectionStatus, timestamp time.Time) error {
	file, openErr := f.openLogFile()
	if openErr != nil {
		return openErr
	}
	defer file.Close()

	logMessage := fmt.Sprintf("[%s] DEVICE: %s STATUS CHANGE: %s -> %s\n", timestamp.Format(time.RFC3339), deviceID, from, to)
	_, writeErr := file.WriteString(logMessage)
	return writeErr
}

func (f *WifiLogger) LogOutageStart(deviceID string, timestamp time.Time) error {
	file, openErr := f.openLogFile()
	if openErr != nil {
		return openErr
	}
	defer file.Close()

	logMessage := fmt.Sprintf("[%s] DEVICE: %s OUTAGE START\n", timestamp.Format(time.RFC3339), deviceID)
	_, writeErr := file.WriteString(logMessage)
	return writeErr
}

func (f *WifiLogger) LogOutageEnd(deviceID string, duration time.Duration, timestamp time.Time) error {
	file, openErr := f.openLogFile()
	if openErr != nil {
		return openErr
	}
	defer file.Close()

	logMessage := fmt.Sprintf("[%s] DEVICE: %s OUTAGE END: LASTED %.0f SECONDS\n", timestamp.Format(time.RFC3339), deviceID, duration.Seconds())
	_, writeErr := file.WriteString(logMessage)
	return writeErr
}

// GetDeviceStats is a stub for the logger, as it does not store stats in a queryable way.
func (f *WifiLogger) GetDeviceStats(deviceID string, since time.Time) (*DeviceData, error) {
	return nil, fmt.Errorf("GetDeviceStats not supported for WifiLogger (logfile only)")
}
