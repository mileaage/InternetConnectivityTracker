package monitor

import (
	"WifiTracker/internals/alerts"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
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

type StorageProvider interface {
	LogConnectivityCheck(deviceID string, success bool, responseTime time.Duration, timestamp time.Time, err error) error
	LogStatusChange(deviceID string, from, to ConnectionStatus, timestamp time.Time) error
	LogOutageStart(deviceID string, timestamp time.Time) error
	LogOutageEnd(deviceID string, duration time.Duration, timestamp time.Time) error
	GetDeviceStats(deviceID string, since time.Time) (*DeviceData, error)
}

type WifiMonitor struct {
	DeviceID      string
	checkInterval time.Duration

	storage StorageProvider

	isRunning  bool
	lastStatus ConnectionStatus

	averageLatency float32
	pingCount      float32

	DataLock sync.RWMutex

	simulateOutage bool
}

type DeviceData struct {
	Online  string
	Latency string
}

var (
	AllDevices   = []*WifiMonitor{}
	devicesMutex sync.RWMutex
)

func New(checkInterval time.Duration, storage StorageProvider) *WifiMonitor {
	monitor := &WifiMonitor{
		DeviceID:       uuid.NewString(),
		checkInterval:  checkInterval,
		storage:        storage,
		isRunning:      false,
		lastStatus:     Inactive,
		simulateOutage: false,
		averageLatency: 0,
		pingCount:      0,
	}

	devicesMutex.Lock()
	AllDevices = append(AllDevices, monitor)
	devicesMutex.Unlock()

	return monitor
}

func (w *WifiMonitor) Start() {
	ticker := time.NewTicker(w.checkInterval)
	defer ticker.Stop()

	w.DataLock.Lock()
	w.lastStatus = Running
	w.isRunning = true
	w.DataLock.Unlock()

	w.logStatusChange(Inactive, Running, time.Now())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	failCount, outageStart := 0, false
	var outageStartTime time.Time

	for {
		select {
		case <-ticker.C:
			down, avgResponse := w.isConnectionDown()

			w.DataLock.Lock()
			w.pingCount++
			w.averageLatency = (w.averageLatency*(w.pingCount-1) + float32(avgResponse.Milliseconds())) / w.pingCount
			w.DataLock.Unlock()

			if down {
				fmt.Printf("Response time: %f seconds\n", avgResponse.Seconds())
				w.logConnectivityCheck(false, avgResponse, ErrConnectionDown)
				failCount++

				if failCount >= 3 {
					if !outageStart {
						outageStartTime = time.Now()
						outageStart = true

						w.DataLock.Lock()
						if w.lastStatus != Down {
							w.logStatusChange(w.lastStatus, Down, time.Now())
							w.lastStatus = Down
						}
						w.DataLock.Unlock()

						w.logOutageStart(time.Now())
					}
				}
			} else {
				failCount = 0
				if outageStart {
					totalDuration := time.Since(outageStartTime)
					w.logOutageEnd(totalDuration, time.Now())
					alerts.SendOutageAlert(totalDuration)
					outageStart = false
				}

				w.logConnectivityCheck(true, avgResponse, nil)

				newStatus := Running
				if avgResponse.Seconds() > 3.0 {
					newStatus = Slow
				}

				w.DataLock.Lock()
				if w.lastStatus != newStatus {
					w.logStatusChange(w.lastStatus, newStatus, time.Now())
					w.lastStatus = newStatus
				}
				w.DataLock.Unlock()
			}

		case <-sigChan:
			fmt.Println("\nMonitoring stopped by user")
			w.DataLock.Lock()
			w.isRunning = false
			w.lastStatus = Inactive
			w.DataLock.Unlock()
			return
		}
	}
}

func (w *WifiMonitor) GetStatus() ConnectionStatus {
	w.DataLock.RLock()
	defer w.DataLock.RUnlock()
	return w.lastStatus
}

func (w *WifiMonitor) GetAverageLatency() float32 {
	w.DataLock.RLock()
	defer w.DataLock.RUnlock()
	return w.averageLatency
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
	w.storage.LogConnectivityCheck(w.DeviceID, success, responseTime, time.Now(), err)
}

func (w *WifiMonitor) logStatusChange(from, to ConnectionStatus, timestamp time.Time) {
	w.storage.LogStatusChange(w.DeviceID, from, to, timestamp)
}

func (w *WifiMonitor) logOutageStart(timestamp time.Time) {
	w.storage.LogOutageStart(w.DeviceID, timestamp)
}

func (w *WifiMonitor) logOutageEnd(duration time.Duration, timestamp time.Time) {
	w.storage.LogOutageEnd(w.DeviceID, duration, timestamp)
}

var defaultTargets = []string{
	"8.8.8.8",
	"1.1.1.1",
	"208.67.222.222",
	"8.8.4.4",
}

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

func GetAllDeviceData() []DeviceData {
	devicesMutex.RLock()
	defer devicesMutex.RUnlock()

	result := make([]DeviceData, 0, len(AllDevices))

	for _, monitor := range AllDevices {
		monitor.DataLock.RLock()
		result = append(result, DeviceData{
			Online:  monitor.lastStatus.String(),
			Latency: fmt.Sprintf("%f", monitor.averageLatency),
		})
		monitor.DataLock.RUnlock()
	}

	return result
}
