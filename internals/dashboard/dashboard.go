package dashboard

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

func MainHandler(w http.ResponseWriter, r *http.Request) {

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
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Read message error: %v", err)
			return
		}

		log.Printf("Received message: %s", string(p))

		if err := conn.WriteMessage(messageType, p); err != nil {
			log.Printf("Write message error: %v", err)
			return
		}
	}

}
