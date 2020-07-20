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
	broadcast: make(chan _Message),
	register:  make(chan User),
}

func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println("Error WebSocketHandler: ", err)
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
