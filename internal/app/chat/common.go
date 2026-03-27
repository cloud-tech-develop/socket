package chat

import (
	"socket/pkg/socket_cors"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	upgrader = socket_cors.Upgrader
)

// Hub gestiona todos los clientes y salas.
type Hub struct {
	mu      sync.RWMutex
	clients map[*websocket.Conn]*Client
	rooms   map[string]*Room
}

// NewHub crea una nueva instancia de Hub.
func NewHub() *Hub {
	return &Hub{
		clients: make(map[*websocket.Conn]*Client),
		rooms:   make(map[string]*Room),
	}
}

// Client representa un usuario conectado.
type Client struct {
	Hub        *Hub
	ID         string
	Connection *websocket.Conn
	Room       *Room
	send       chan Message
}

// Room representa una sala de chat.
type Room struct {
	ID      string
	Clients map[*Client]bool
	mu      sync.RWMutex
}

// Message representa un mensaje de chat.
type Message struct {
	Sender   int64     `json:"sender"`
	Content  any       `json:"content"`
	DateTime time.Time `json:"dateTime"`
}
