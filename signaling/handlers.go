package signaling

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var RManager = RoomManager{
	rooms:     make(map[string]*Room),
	forward: make(chan _Message),
	register:  make(chan User),
	unregister: make(chan Unregister),
	admission: make(chan Admit),
}

func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { 
		origin := r.Header["Origin"]
		log.Printf("Origin: %v", origin)
		if origin[0] == "http://localhost:8000" || origin[0] == "https://proximo.netlify.app" {
			return true
		}
		return false 
	}
	ws, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Printf("Error WebSocketHandler: %v", err)
		return
	}
	connection := &Connection{
		ws:   ws,
		send: make(chan []byte, 256),
	}
	log.Printf("Ws Connection Established: %v", connection)
	// make connection here and start readMessage in a thread
	go connection.readMessage()
	go connection.writeMessage()
}
