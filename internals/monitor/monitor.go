package monitor

import (
	"WifiTracker/internals/alerts"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

var ErrConnectionDown = errors.New("connection is down")

type ConnectionStatus int

const (
	Running ConnectionStatus = iota
	Slow
	Down
	Inactive
)

func (c ConnectionStatus) String() string {
	switch c {
	case Running:
		return "RUNNING"
	case Slow:
		return "SLOW"
	case Down:
		return "DOWN"
	case Inactive:
		return "INACTIVE"
	default:
		return "UNKNOWN"
	}
}

type WifiMonitor struct {
	checkInterval time.Duration

	// dependencies
	logger WifiLogger

	// state
	isRunning  bool
	lastStatus ConnectionStatus

	// mock
	simulateOutage bool
}

func New(checkInterval time.Duration, logger *WifiLogger) *WifiMonitor {
	return &WifiMonitor{
		checkInterval:  checkInterval,
		logger:         *logger,
		isRunning:      false,
		lastStatus:     Inactive,
		simulateOutage: false,
	}
}

func (w *WifiMonitor) Start() {
	ticker := time.NewTicker(w.checkInterval)
	defer ticker.Stop()

	w.lastStatus = Running
	w.logStatusChange(Inactive, Running, time.Now())

	// handle ctrl + c
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	failCount, outageStart := 0, false
	var outageStartTime time.Time

	for {
		select {
		case <-ticker.C:
			down, avgResponse := w.isConnectionDown()

			if down {
				fmt.Printf("Response time: %f seconds\n", avgResponse.Seconds())
				w.logConnectivityCheck(false, avgResponse, ErrConnectionDown)
				failCount++

				if failCount >= 3 {
					if !outageStart {
						outageStartTime = time.Now()
						outageStart = true

						w.logOutageStart(time.Now())
					}
				}
			} else {
				// good connection
				failCount = 0
				if outageStart {
					// Calculate total outage duration
					totalDuration := time.Since(outageStartTime)
					w.logOutageEnd(totalDuration, time.Now())
					alerts.SendOutageAlert(time.Since(outageStartTime))
					outageStart = false
				}

				// log success too
				w.logConnectivityCheck(true, avgResponse, nil)

				// now we check the speed
				if avgResponse.Seconds() > 3.0 {
					if w.lastStatus != Slow {
						w.logStatusChange(w.lastStatus, Slow, time.Now())
						w.lastStatus = Slow
					} else {
						w.logStatusChange(w.lastStatus, Running, time.Now())
						w.lastStatus = Running
					}
				}
			}

		case <-sigChan:
			fmt.Println("\nMonitoring stopped by user")
			return
		}
	}
}

func (w *WifiMonitor) ping(target string) (bool, time.Duration) {
	start := time.Now()

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("ping", "-n", "1", target)
	case "darwin", "linux", "freebsd", "openbsd", "netbsd":
		cmd = exec.Command("ping", "-c", "1", target)
	default:
		// fall back for unknown os
		cmd = exec.Command("ping", "-c", "1", target)
	}
	err := cmd.Run()
	responseTime := time.Since(start)

	if err != nil {
		return false, responseTime
	}

	return true, responseTime
}

func (w *WifiMonitor) logConnectivityCheck(success bool, responseTime time.Duration, err error) {
	w.logger.LogConnectivityCheck(success, responseTime, err)
}

func (w *WifiMonitor) logStatusChange(from, to ConnectionStatus, timestamp time.Time) {
	w.logger.LogStatusChange(from, to, timestamp)
}

func (w *WifiMonitor) logOutageStart(timestamp time.Time) {
	w.logger.LogOutageStart(timestamp)
}

func (w *WifiMonitor) logOutageEnd(duration time.Duration, timestamp time.Time) {
	w.logger.LogOutageEnd(duration, timestamp)
}

var defaultTargets = []string{
	"8.8.8.8",        // Google DNS
	"1.1.1.1",        // Cloudflare DNS
	"208.67.222.222", // OpenDNS
	"8.8.4.4",        // Google DNS secondary
}

// Consider connection down only if 3+ targets fail
func (w *WifiMonitor) isConnectionDown() (bool, time.Duration) {
	failures := 0
	totalDuration := time.Duration(0)

	for _, target := range defaultTargets {
		ok, duration := w.ping(target)
		if !ok {
			failures++
		}
		totalDuration += duration
	}

	avgDuration := totalDuration / time.Duration(len(defaultTargets))
	isDown := failures >= 3

	return isDown, avgDuration
}
