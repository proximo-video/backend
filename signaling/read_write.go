package signaling

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next message from the peer.
	readWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than readWait(pongWait).
	pingPeriod = (readWait * 9) / 10
)

// readMessage will constantly read message from the websocket connection
func (connection *Connection) readMessage() {
	// set maximum time limit for reading a messgage
	connection.ws.SetReadDeadline(time.Now().Add(readWait))
	connection.ws.SetPongHandler(func(string) error { connection.ws.SetReadDeadline(time.Now().Add(readWait)); return nil })
	user := User{connection: connection}
	defer func() {
		log.Printf("Closing Connection in readMessage: %v", connection.userId)
		RManager.unregister <- Unregister{user: user, action: SELF}
		connection.ws.Close()
	}()
	for {
		_, byteMsg, err := connection.ws.ReadMessage()
		if err != nil {
			log.Printf("error: %v", err)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v", err)
			}
			break
		}
		var msg Message
		err = json.Unmarshal(byteMsg, &msg)
		if err != nil {
			log.Printf("error in unmarshalling in readMessage: %v", err)
			break
		}
		// take suitable actions
		switch msg.Action {
		case START:
			//TODO: Checks
			// 1. Rooms should exist in database
			// 2. If user is the owner then room should belong to him
			// set user id
			connection.userId = msg.UserId
			// data in messgae will be room id
			user.roomId = msg.Data.(string)
			// only owner of a room can start a meeting
			user.isOwner = true
			RManager.register <- user
			log.Printf("Start from user: %v", msg.UserId)
			// handle one more thing sending the reply back
			// reply should be handled after the registration so handle in room_managers
		case JOIN:
			//TODO: Checks
			// 1. Rooms should exist in database
			// set user id
			connection.userId = msg.UserId
			// data in messgae will be room id
			user.roomId = msg.Data.(string)
			// only owner of a room can start a meeting
			user.isOwner = false
			RManager.register <- user
			log.Printf("Join from user: %v", msg.UserId)
		case END:
			// handle deregistration
			// only applicable when the requester is the owner of the room
			// check if user is the owner
			if user.isOwner {
				// iterate though all users and remove them one by one
				RManager.unregister <- Unregister{user: user, action: ALL}
			} else {
				log.Println("User is not the owner of the room END not applicable")
				// TODO: send some reply to user
			}
		case LEAVE:
			// just send user to leave
			RManager.unregister <- Unregister{user: user, action: SELF}
		case MESSAGE:
			// log.Printf("broad from user: %v", user.connection.userId)
			if user.roomId != "" {
				// log.Printf("Broadcast from user: %v for room %v", user.connection.userId, user.roomId)
				broadcastMess := _Message{
					ws:      connection.ws,
					message: msg,
					roomId:  user.roomId,
				}
				log.Printf("Broadcast from user send to broadcast channel: %v", user.connection.userId)
				RManager.broadcast <- broadcastMess
			} else {
				log.Printf("Error in broadcast message: %v", err)
				// TODO: send some reply of error
			}
		}
	}
	log.Printf("For loop break")
}

// write writes a message with the given message type and payload.
func (c *Connection) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

func (connection *Connection) writeMessage() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		connection.ws.Close()
	}()
	for {
		select {
		case message, ok := <-connection.send:
			if !ok {
				log.Printf("Closing connection in writeMessage")
				connection.write(websocket.CloseMessage, []byte{})
				return
			}
			if err := connection.write(websocket.TextMessage, message); err != nil {
				log.Printf("Error in writeMessage: %v", err)
				return
			}
		case <-ticker.C:
			if err := connection.write(websocket.PingMessage, []byte{}); err != nil {
				log.Printf("Error in writeMessage during Ping: %v", err)
				return
			}
		}
	}
}
