package signaling

import (
	"WebRTCConf/auth"
	"WebRTCConf/database"
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
			if user.isOwner {
				dbRoom, err := database.GetUserRoom(auth.Ctx, auth.Client, user.connection.userId, user.roomId)
				if err != nil {
					log.Printf("Register: error: %v", err)
				} else {
					room, ok := r.rooms[user.roomId]
					if !ok { // case room not found
						room = &Room{ // create new room
							roomId:    user.roomId,
							isLocked:  dbRoom.IsLocked,
							users:     make(map[*Connection]bool),
							waitUsers: make(map[string]*Connection),
						}
						room.users[user.connection] = true
						room.owner = user.connection
						r.rooms[user.roomId] = room
						log.Printf("Registered Owner: %v and Created room: %v", user.connection.userId, user.roomId)
					} else {
						if _, ok := room.users[user.connection]; !ok {
							room.users[user.connection] = true
							room.owner = user.connection
						}
						// send ready to all users
						for uc := range room.users {
							marshalled, err := json.Marshal(Message{
								Action: READY,
								To:     uc.userId,
								From:   user.connection.userId,
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
						// if room is locked then send ENTER request
						if dbRoom.IsLocked {
							for ucId, _ := range room.waitUsers {
								marshalled, err := json.Marshal(Message{
									Action: ENTER,
									To:     room.owner.userId,
									From:   ucId,
								})
								if err != nil {
									log.Printf("Register: Marshalling Error in sending ENTER action: %v", err)
									continue
								}
								room.owner.send <- marshalled
							}
						}
					}
				}
			} else {
				dbRoom, err := database.GetRoom(auth.Ctx, auth.Client, user.roomId)
				if err != nil {
					log.Printf("Register: user: %v, Error in GetRoom: %v", user.connection.userId, err)
				} else {
					if room, ok := r.rooms[user.roomId]; !ok {
						room = &Room{ // create new room
							roomId:    user.roomId,
							isLocked:  dbRoom.IsLocked,
							users:     make(map[*Connection]bool),
							waitUsers: make(map[string]*Connection),
						}
						r.rooms[user.roomId] = room
					}
					room := r.rooms[user.roomId]
					if room.isLocked { // room is Locked
						// put user in waitUsers queue
						room.waitUsers[user.connection.userId] = user.connection
						// if owner of the room is present then send ENTER request
						if room.owner != nil {
							marshalled, err := json.Marshal(Message{
								Action: ENTER,
								To:     room.owner.userId,
								From:   user.connection.userId,
							})
							if err != nil {
								log.Printf("Register: Marshalling Error in sending ENTER action: %v", err)
								continue
							}
							room.owner.send <- marshalled
						}
					} else { // room is not Locked
						room.users[user.connection] = true
						// send ready to all users
						for uc := range room.users {
							marshalled, err := json.Marshal(Message{
								Action: READY,
								To:     uc.userId,
								From:   user.connection.userId,
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
