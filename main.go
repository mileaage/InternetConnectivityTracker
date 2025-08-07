package main

import (
	"log"
	"time"

	"WifiTracker/internals/dashboard"
	"WifiTracker/internals/monitor"
)

func main() {
	myLogger, err := monitor.NewWifiLogger("log.txt")
	if err != nil {
		panic(err)
	}

	go func() {
		log.Println("starting server (please don't block)")
		dashboard.StartDashboard()
	}()

	log.Printf("starting monitor")
	myMonitor := monitor.New(time.Second, myLogger)
	myMonitor.Start()

}
