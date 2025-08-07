package main

import (
	"net/http"

	"github.com/a-h/templ"
)

// "WifiTracker/internals/monitor"

func main() {
	// myLogger, err := monitor.NewWifiLogger("log.txt")
	// if err != nil {
	// 	panic(err)
	// }

	// myMonitor := monitor.New(time.Second, myLogger)
	// myMonitor.Start()

	// start the dashboard

	// start the websocket server separately
	component := headerTemplate("WifiConnect")

	http.Handle("/", templ.Handler(component))
	http.ListenAndServe(":3000", nil)
}
