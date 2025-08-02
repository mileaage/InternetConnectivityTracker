// In monitor/logger.go
package monitor

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Logger interface {
	LogConnectivityCheck(target string, success bool, responseTime time.Duration, err error)
	LogStatusChange(from, to ConnectionStatus, timestamp time.Time)
	LogOutageStart(target string, timestamp time.Time)
	LogOutageEnd(target string, duration time.Duration, timestamp time.Time)
}

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

func (f *WifiLogger) LogConnectivityCheck(success bool, responseTime time.Duration, err error) {
	file, openErr := f.openLogFile()
	if openErr != nil {
		fmt.Fprintf(os.Stderr, "Failed to open log file for connectivity check: %v\n", openErr)
		return
	}
	defer file.Close()

	var logMessage string
	if success {
		logMessage = fmt.Sprintf("DATE: %s CONNECTED IN %.2f SECONDS\n", time.Now().Format("1/2/2006"), responseTime.Seconds())
	} else {
		logMessage = fmt.Sprintf("DATE: %s FAILED TO CONNECT IN %.2f SECONDS, ERROR: %v\n", time.Now().Format("1/2/2006"), responseTime.Seconds(), err)
	}

	if _, writeErr := file.WriteString(logMessage); writeErr != nil {
		fmt.Fprintf(os.Stderr, "Failed to write connectivity check log: %v\n", writeErr)
	}
}

func (f *WifiLogger) LogStatusChange(from, to ConnectionStatus, timestamp time.Time) {
	file, openErr := f.openLogFile()
	if openErr != nil {
		fmt.Fprintf(os.Stderr, "Failed to open log file for status change: %v\n", openErr)
		return
	}
	defer file.Close()

	logMessage := fmt.Sprintf("STATUS CHANGE %s: %s -> %s\n", timestamp.Format("1/2/2006"), from, to)

	if _, writeErr := file.WriteString(logMessage); writeErr != nil {
		fmt.Fprintf(os.Stderr, "Failed to write status change log: %v\n", writeErr)
	}
}

func (f *WifiLogger) LogOutageStart(timestamp time.Time) {
	file, openErr := f.openLogFile()
	if openErr != nil {
		fmt.Fprintf(os.Stderr, "Failed to open log file: %+v", openErr)
		return
	}

	defer file.Close()

	logMessage := fmt.Sprintf("OUTAGE START %s", timestamp.Format("1/2/2006"))

	if _, writeErr := file.WriteString(logMessage); writeErr != nil {
		fmt.Fprintf(os.Stderr, "Failed to write outage start: %v\n", writeErr)
	}
}

func (f *WifiLogger) LogOutageEnd(duration time.Duration, timestamp time.Time) {
	file, openErr := f.openLogFile()
	if openErr != nil {
		fmt.Fprintf(os.Stderr, "Failed to open log file for outage end: %v\n", openErr)
		return
	}
	defer file.Close()

	logMessage := fmt.Sprintf("OUTAGE END %s: LASTED %.0f SECONDS\n", timestamp.Format("1/2/2006"), duration.Seconds())

	if _, writeErr := file.WriteString(logMessage); writeErr != nil {
		fmt.Fprintf(os.Stderr, "Failed to write outage end log: %v\n", writeErr)
	}
}
