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

	storage, err := db.NewDatabaseStorage("downtimedata.db")
	if err != nil {
		panic(err)
	}

	storage.LogOutageStart("bbed78db-4aa8-46bc-930e-e689aabf5eb0", time.Now())
	storage.LogOutageEnd("bbed78db-4aa8-46bc-930e-e689aabf5eb0", time.Minute, time.Now())

	// create the database connection
	downtimeStorage, err := db.NewDatabaseStorage("downtimedata.db")
	if err != nil {
		log.Fatalf("Error opening database: %v\n", err)
	}

	for {

		deviceData := monitor.GetAllDeviceData()
		dayDowntime := DowntimeErrorCheck(downtimeStorage, util.OneDayAgo())
		weekDowntime := DowntimeErrorCheck(downtimeStorage, util.OneWeekAgo())
		monthDowntime := DowntimeErrorCheck(downtimeStorage, util.OneMonthAgo())

		// log.Printf("Data:\nDay: %+v\nWeek: %+v\nMonth: %+v\n", dayDowntime, weekDowntime, monthDowntime)

		if len(deviceData) == 0 {
			continue
		}

		fmt.Printf("%+v\n", deviceData)

		// experimental for now
		firstValue := deviceData[0]

		valuesOver := struct {
			Online         string
			Latency        string
			DayDowntimes   []db.DowntimeEvent
			WeekDowntimes  []db.DowntimeEvent
			MonthDowntimes []db.DowntimeEvent
		}{
			Online:         firstValue.Online,
			Latency:        firstValue.Latency,
			DayDowntimes:   dayDowntime,
			WeekDowntimes:  weekDowntime,
			MonthDowntimes: monthDowntime,
		}

		log.Printf("%+v\n", valuesOver)

		if err := conn.WriteJSON(valuesOver); err != nil {
			log.Fatalf("Error Parsing JSON: %v", deviceData)
		}

		time.Sleep(time.Second)
	}

}

func DowntimeErrorCheck(db *db.DatabaseStorage, ts time.Time) []db.DowntimeEvent {
	result, err := db.GetDowntimes(ts)
	if err != nil {
		log.Printf("Error fetching downtimes: %v", err)
		return nil
	}
	return result
}

func StartDashboard() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	http.HandleFunc("/ws", WebsocketHandler)

	http.ListenAndServe("localhost:8080", nil)
}
