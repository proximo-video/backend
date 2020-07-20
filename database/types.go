package database

type Room struct {
	RoomId   string `json:"room_id,omitempty"`
	IsLocked bool   `json:"is_locked,string,omitempty"`
}

// User is a json-serializable type.
type User struct {
	Id    string `json:"id,omitempty"`
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
	Rooms []Room `json:"rooms,omitempty"`
}
