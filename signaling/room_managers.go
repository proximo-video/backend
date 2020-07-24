package signaling

import (
	"encoding/json"
	"log"
)

func (r *RoomManager) deleteUser(connection *Connection, room *Room) {
	log.Printf("Deleting user: %s from room: %s", connection.userId, room.roomId)
	// remove user from the room
	delete(room.users, connection)
	// if user is owner then remove from owner field
	if connection == room.owner {
		room.owner = nil
	}
	// close channel
	close(connection.send)
	// if no more users are left then delete the room
	if len(room.users) == 0 {
		log.Printf("Deleting room: %v", room.roomId)
		delete(r.rooms, room.roomId)
	}
}

func (r *RoomManager) HandleChannels() {
	for {
		select {
		case user := <-r.register:
			room, ok := r.rooms[user.roomId]
			if !ok { // case room not found
				// create new room
				room = &Room{
					roomId:   user.roomId,
					isLocked: false, // handle later
					users:    make(map[*Connection]bool),
				}
				room.users[user.connection] = true
				// if this connection is the owner of the room register user as owner
				if user.isOwner {
					room.owner = user.connection
				}
				r.rooms[user.roomId] = room
				log.Printf("Registered first User: %v", user.connection.userId)
			} else {
				// TODO: if the user already exists in the room then don't register
				// if room exists then handle
				// Send READY signal to owner and WAIT signal to other member
				if _, ok := room.users[user.connection]; !ok {
					room.users[user.connection] = true
				}
				if user.isOwner {
					room.owner = user.connection
				}
				// send READY to all other users
				log.Printf("Registered %v User: %v", len(room.users), user.connection.userId)
				for uc := range room.users {
					marshalled, err := json.Marshal(Message{
						Action: READY,
						To: uc.userId,
						From: user.connection.userId,
					})
					if err != nil {
						log.Printf("Marshalling Error in Register User: %v", err)
						continue
					}
					// log.Printf("Message for conn: %v sender %v", uc.userId, mess.message.UserId)
					// don't send message to sender again
					if uc.ws != user.connection.ws {
						log.Printf("Sending message to user %v from user %v", uc.userId, user.connection.userId)
						// send to the user channel
						uc.send <- marshalled
					} 
				}
			}
		case unregis := <-r.unregister:
			user := unregis.user
			room, ok := r.rooms[user.roomId]
			if !ok {
				log.Printf("Room %s Not found for user: %s", user.roomId, user.connection.userId)
			} else {
				if unregis.action == ALL {
					log.Printf("Delete All Users %s in room %s", user.connection.userId, user.roomId)
					for uc := range room.users {
						r.deleteUser(uc, room)
					}
				} else {
					log.Printf("Delete Self User %s in room %s", user.connection.userId, user.roomId)
					if _, ok := room.users[user.connection]; ok {
						r.deleteUser(user.connection, room)
					} else {
						log.Printf("User %s not present in room %s", user.connection.userId, user.roomId)
					}
				}
			}
		case mess := <-r.forward:
			if room, ok := r.rooms[mess.roomId]; ok {
				// get marshal
				marshalled, err := json.Marshal(mess.message)
				if err != nil {
					log.Printf("error in marshalling in forward: %v", err)
				} else {
					// loop through all users and forward message to the destination connection only
					for uc := range room.users {
						// log.Printf("Message for conn: %v sender %v", uc.userId, mess.message.UserId)
						// don't send message to sender again
						if uc.ws != mess.ws && mess.message.To == uc.userId {
							log.Printf("Sending forward to user %v from user %v", uc.userId, mess.message.From)
							// send to the user channel
							uc.send <- marshalled
							break
						}
					}
				}
			}
		case admitMess := <-r.admission:
			if room, ok := r.rooms[admitMess.roomId]; ok {
				if admitMess.action == APPROVE {
					conn, ok := room.waitUsers[admitMess.userId]
					if ok {
						room.users[conn] = true
						delete(room.waitUsers, admitMess.userId)
					} else {
						log.Printf("Error in Admission channel. User: %s not present in waitUsers of room: %s", admitMess.userId, room.roomId)
					}
				}
			}
		}
	}
}
