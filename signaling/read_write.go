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
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("Error in ReadMessage of readMessage: %v", err)
			}
			break
		}
		var msg Message
		err = json.Unmarshal(byteMsg, &msg)
		if err != nil {
			log.Printf("Error in unmarshalling in readMessage: %v", err)
			break
		}
		// take suitable actions
		switch msg.Action {
		case START:
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
				log.Printf("User: %s is not the owner of the room: %s so END is not applicable", user.connection.userId, user.roomId)
				// TODO: send some reply to user
			}
		case LEAVE:
			// just send user to leave
			RManager.unregister <- Unregister{user: user, action: SELF}
		case MESSAGE:
			if user.roomId != "" &&  msg.From != "" && msg.To != "" && msg.From != msg.To{
				frowardMess := _Message{
					ws: connection.ws,
					message: msg,
					roomId: user.roomId,
				}
				RManager.forward <- frowardMess
				log.Printf("Forward message from user: %v to user: %v", msg.From, msg.To)
			} else {
				log.Printf("Invalid RoomId: %v or msg.From: %v or msg.To: %v in MESSAGE", user.roomId, msg.From, msg.To)
			}
		case APPROVE, REJECT:
			if user.roomId != "" && msg.To != "" {
				admitMess := Admit{
					userId: msg.To,
					action: msg.Action,
					roomId: user.roomId,
				}
				RManager.admission <- admitMess
				log.Printf("Approve entrance of user: %s by owner: %s for room: %s", msg.To, user.connection.userId, user.roomId)
			} else {
				log.Printf("Invalid RoomId: %v or msg.To: %v in APPROVE", user.roomId, msg.To)
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
