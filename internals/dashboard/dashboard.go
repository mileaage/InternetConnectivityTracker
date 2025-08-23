package dashboard

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"WifiTracker/internals/db"
	"WifiTracker/internals/monitor"
	"WifiTracker/util"

	"github.com/gorilla/websocket"
)

// TODO:
// switch out templates for just text/template since it's more general purpose
// complete dashboard, add statistics, cost impact
// add multiple protocols
// scrape potential downtimes and attribute them
// add measures of seveity
// maintenance windows
// custom webhooks perhaps
// group related notifications if in a certain time period
// have messages ready for teams / groupchats after down time
// metrics export feature

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // allow all
	},
}

func WebsocketHandler(w http.ResponseWriter, r *http.Request) {
	// upgrade
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		http.Error(w, "WebSocket upgrade failed", http.StatusBadRequest)
		return
	}
	defer conn.Close()

	log.Println("WebSocket connection established")

	for {

		deviceData := monitor.GetAllDeviceData()
		dayDowntime := DowntimeErrorCheck(util.OneDayAgo())
		weekDowntime := DowntimeErrorCheck(util.OneWeekAgo())
		monthDowntime := DowntimeErrorCheck(util.OneMonthAgo())

		// TODO: Add proper ranging

		log.Printf("Data:\nDay: %+v\nWeek: %+v\nMonth: %+v\n", dayDowntime, weekDowntime, monthDowntime)

		if err != nil {
			log.Printf("Error fetching any downtimes: %v", err)
			http.Error(w, "Error fetching downtimes", http.StatusBadRequest)
		}

		if len(deviceData) == 0 {
			continue
		}

		fmt.Printf("%+v\n", deviceData)

		// experimental for now
		firstValue := deviceData[0]

		if err := conn.WriteJSON(firstValue); err != nil {
			log.Fatalf("Error Parsing JSON: %v", deviceData)
		}

		time.Sleep(time.Second)
	}

}

func DowntimeErrorCheck(ts time.Time) []db.DowntimeEvent {
	results, err := db.GetDowntimes(ts)
	if err != nil {
		log.Printf("Error fetching downtimes")
		return nil
	}

	return results
}

func StartDashboard() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	http.HandleFunc("/ws", WebsocketHandler)

	http.ListenAndServe("localhost:8080", nil)
}
