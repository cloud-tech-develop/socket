package chat

import (
	"socket/pkg/socket_cors"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	upgrader = socket_cors.Upgrader
	clients  = make(map[*websocket.Conn]*Client)
	rooms    = make(map[string]*Room)
	mu       = &sync.Mutex{}
)

type Client struct {
	ID         string
	Connection *websocket.Conn
	Room       *Room
}

type Room struct {
	ID      string
	Clients map[*Client]bool
	mu      sync.Mutex
}

type Message struct {
	Sender   int64     `json:"sender"`
	Content  string    `json:"content"`
	DateTime time.Time `json:"dateTime"`
}

type RestProxy struct {
	Upgrader websocket.Upgrader
	Clients  map[*websocket.Conn]*Client
	Rooms    map[string]*Room
	Mu       *sync.Mutex
}
