package signaling

import "errors"
import "fmt"

var BadMessage error = errors.New("Bad message. Invalid room id or to user or from user or user id.")
var NotOwner error = errors.New("User is not the owner of room.")
var NotFound error = errors.New("Room or user not found.")
//var UserAlreadyExists = errors.New("User already present in room")
//var UserNotPresent = errors.New("User not present")
func UserAlreadyPresent(userId string, roomId string) error  {
	return fmt.Errorf("User: %v already present in room: %v.", userId, roomId)
}

func UserNotPresent(userId string, roomId string) error  {
	return fmt.Errorf("User: %v not present in room: %v.", userId, roomId)
}

func RoomNotFound(roomId string) error  {
	return fmt.Errorf("Room not found: %v.", roomId)
}
