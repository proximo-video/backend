package signaling

import (
	"github.com/gorilla/websocket"
)

// Message type for incoming messages from the client
type Message struct {
	Action      string      `json:"action,omitempty"`
	UserId      string      `json:"id,omitempty"`
	Data        interface{} `json:"data,omitempty"`
	Type        string      `json:"type,omitempty"`
	To          string      `json:"to,omitempty"`
	From        string      `json:"from,omitempty"`
	DisplayName string      `json:"display_name,omitempty"`
}

// Message type for internal usage only
type _Message struct {
	ws      *websocket.Conn
	message Message
	roomId  string
}

// Connection type will store all infomation for about a connection
type Connection struct {
	ws     *websocket.Conn
	userId string
	send   chan []byte
}

type User struct {
	connection *Connection
	roomId     string
	isOwner    bool
}

type Room struct {
	roomId    string
	isLocked  bool
	owner     *Connection
	users     map[*Connection]bool
	waitUsers map[string]*Connection
}

type Unregister struct {
	user   User
	action string
}

type Admit struct {
	userId string
	action string
	roomId string
}

type RoomManager struct {
	rooms map[string]*Room

	// forward channel to handle forwarding messages
	forward chan _Message

	// register channel to handle registration / meeting start request from clients
	register chan User

	// unregister channel to handle leave / end meeting request from clients
	unregister chan Unregister

	// admission channel to handle all admission / rejection replies from the owner of the room
	admission chan Admit
}
