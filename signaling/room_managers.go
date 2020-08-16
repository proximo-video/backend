package signaling

import (
	"WebRTCConf/auth"
	"WebRTCConf/database"
	"encoding/json"
	"log"
)

func (c *Connection) SendError(err error) {
	marshalled, err := json.Marshal(Message{
		Action: ERROR,
		UserId: c.userId,
		Data: err.Error(),
		To: c.userId,
	})
	if err == nil {
		c.send <- marshalled
	} else {
		log.Printf("SendError: Error in marshaling: %v", err)
	}
}

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
		// Before deleting room close all connections in waitUsers queue also
		for ucId, uc := range room.waitUsers {
			close(uc.send)
			delete(room.waitUsers, ucId)
		}
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
					user.connection.SendError(err)
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
							// check if user already exists: don't re-enter user
						if _, ok := room.users[user.connection]; !ok {
							room.users[user.connection] = true
							room.owner = user.connection
							
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
							// if room is locked then send PERMIT request to owner from all waiting Users
							if dbRoom.IsLocked {
								for ucId, uc := range room.waitUsers {
									marshalled, err := json.Marshal(Message{
										Action: PERMIT,
										To:     room.owner.userId,
										From:   ucId,
										DisplayName: uc.displayName,
									})
									if err != nil {
										log.Printf("Register: Marshalling Error in sending PERMIT action: %v", err)
										continue
									}
									room.owner.send <- marshalled
								}
							}
						} else {
							log.Printf("1: User: %s already present in room: %s", user.connection.userId, user.roomId)
						}
					}
				}
			} else {
				dbRoom, err := database.GetRoom(auth.Ctx, auth.Client, user.roomId)
				log.Printf("got dbRoom: %v",dbRoom)
				if err != nil {
					log.Printf("Register: user: %v, Error in GetRoom: %v", user.connection.userId, err)
					user.connection.SendError(err)
				} else {
					if room, ok := r.rooms[user.roomId]; !ok {
						log.Printf("Create new room %v in server from user: %v", user.roomId, user.connection.userId)
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
						_, ok := room.waitUsers[user.connection.userId]
						_, ok1 := room.users[user.connection]
						// put user in waitUsers queue: only if its not already present in waitUsers and users
						if !ok && !ok1 {
							room.waitUsers[user.connection.userId] = user.connection
							
							// if owner of the room is present then send PERMIT request
							if room.owner != nil {
								marshalled, err := json.Marshal(Message{
									Action: PERMIT,
									To:     room.owner.userId,
									From:   user.connection.userId,
									DisplayName: user.connection.displayName,
								})
								if err != nil {
									log.Printf("Register: Marshalling Error in sending PERMIT action: %v", err)
									continue
								} else {
									room.owner.send <- marshalled
								}
							}
							marshalled, err := json.Marshal(Message{
								Action: WAIT,
								To:     user.connection.userId,
								From:   "server",
							})
							if err == nil {
								user.connection.send <- marshalled
							} else {
								log.Printf("Register: Marshalling Error in sending WAIT action: %v", err)
							}
						} else {
							log.Printf("2: User: %s already present in room: %s", user.connection.userId, user.roomId)
						}
					} else { // room is not Locked
						log.Printf("Room is not locked: %v", room.roomId)
						if _, ok := room.users[user.connection]; !ok {
							log.Printf("registered new user in room: %v", user.connection.userId)
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
									log.Printf("Sending ready to user %v from user %v", uc.userId, user.connection.userId)
									// send to the user channel
									uc.send <- marshalled
								}
							}
						} else {
							log.Printf("3: User: %s already present in room: %s", user.connection.userId, user.roomId)
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
					} else if _, ok := room.waitUsers[user.connection.userId]; ok {
						// TODO: Check if there are no more users in waitUsers and users then delete the room
						close(user.connection.send)
						delete(room.waitUsers, user.connection.userId)
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
						marshalled, err := json.Marshal(Message{
							Action: APPROVE,
							To:     admitMess.userId,
							From:   room.owner.userId,
						});
						if err != nil {
							log.Printf("Admission channel: Marshalling Error in REJECT: %v", err)
						} else {
							conn.send <- marshalled
							// send ready to all users
							for uc := range room.users {
								marshalled, err := json.Marshal(Message{
									Action: READY,
									To:     uc.userId,
									From:   conn.userId,
								})
								if err != nil {
									log.Printf("Admission channel: Marshalling Error: %v", err)
									continue
								}
								// log.Printf("Message for conn: %v sender %v", uc.userId, mess.message.UserId)
								// don't send message to sender again
								if uc.ws != conn.ws {
									log.Printf("Admission channel: Sending message to user %v from user %v", uc.userId, conn.userId)
									// send to the user channel
									uc.send <- marshalled
								}
							}
						}
						// TODO: Send some reply to accepted user.
					} else {
						log.Printf("Admission channel: User: %s not present in waitUsers of room: %s", admitMess.userId, room.roomId)
					}
				} else { // User has been denied permission to enter
					conn, ok := room.waitUsers[admitMess.userId]
					if ok {
						marshalled, err := json.Marshal(Message{
							Action: REJECT,
							To:     admitMess.userId,
							From:   room.owner.userId,
						})
						if err != nil {
							log.Printf("Admission channel: Marshalling Error in REJECT: %v", err)
						} else {
							conn.send <- marshalled
							close(conn.send)
							delete(room.waitUsers, admitMess.userId)
						}
					} else {
						log.Printf("Admission channel REJECT: User: %s not present in waitUsers of room: %s", admitMess.userId, room.roomId)
					}
				}
			}
		}
	}
}
