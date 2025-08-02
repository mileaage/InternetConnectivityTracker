package main

import (
	"time"

	"WifiTracker/internals/monitor"
)

func main() {
	myLogger, err := monitor.NewWifiLogger("log.txt")
	if err != nil {
		panic(err)
	}

	myMonitor := monitor.New(time.Second, myLogger)
	myMonitor.Start()
}
