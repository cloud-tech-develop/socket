package chat

import (
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

func (h *Hub) HandleWebSocket(c echo.Context) error {
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Printf("Error al actualizar la conexión WebSocket: %s", err)
		return echo.NewHTTPError(400, "Error al actualizar la conexión WebSocket")
	}

	clientID := c.QueryParam("id")
	if clientID == "" {
		clientID = generateUniqueID()
	}

	roomID := c.QueryParam("room")
	if roomID == "" {
		roomID = "public"
	}

	client := &Client{
		Hub:        h,
		ID:         clientID,
		Connection: conn,
		send:       make(chan Message, 256),
	}

	if err := h.joinRoom(client, roomID); err != nil {
		log.Printf("Error al unirse a la sala: %s", err)
		conn.Close()
		return echo.NewHTTPError(500, "Error al unirse a la sala")
	}

	log.Printf("Cliente %s conectado a la sala %s\n", client.ID, roomID)

	// Iniciar goroutines para lectura y escritura
	go client.writePump()
	go client.readPump(roomID)

	return nil
}

func (h *Hub) joinRoom(client *Client, roomID string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.rooms[roomID]; !ok {
		h.rooms[roomID] = &Room{
			ID:      roomID,
			Clients: make(map[*Client]bool),
		}
	}

	client.Room = h.rooms[roomID]
	client.Room.mu.Lock()
	client.Room.Clients[client] = true
	client.Room.mu.Unlock()

	h.clients[client.Connection] = client
	return nil
}

func (h *Hub) leaveRoom(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if client.Room != nil {
		client.Room.mu.Lock()
		delete(client.Room.Clients, client)
		roomEmpty := len(client.Room.Clients) == 0
		client.Room.mu.Unlock()

		if roomEmpty {
			delete(h.rooms, client.Room.ID)
		}
	}
	delete(h.clients, client.Connection)
}

func (h *Hub) Broadcast(roomID string, message Message) {
	h.mu.RLock()
	room, ok := h.rooms[roomID]
	h.mu.RUnlock()

	if !ok {
		return
	}

	message.DateTime = time.Now()
	room.mu.RLock()
	defer room.mu.RUnlock()

	for client := range room.Clients {
		select {
		case client.send <- message:
		default:
			close(client.send)
			h.leaveRoom(client)
		}
	}
}

func (c *Client) readPump(roomID string) {
	defer func() {
		c.Hub.leaveRoom(c)
		c.Connection.Close()
	}()

	c.Connection.SetReadLimit(maxMessageSize)
	c.Connection.SetReadDeadline(time.Now().Add(pongWait))
	c.Connection.SetPongHandler(func(string) error { c.Connection.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		var message Message
		err := c.Connection.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		log.Printf("Mensaje recibido de %s en la sala '%s': %+v\n", c.ID, roomID, message)
		c.Hub.Broadcast(roomID, message)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Connection.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.Connection.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// El Hub cerró el canal.
				c.Connection.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Connection.WriteJSON(message); err != nil {
				return
			}

		case <-ticker.C:
			c.Connection.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Connection.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func generateUniqueID() string {
	return uuid.New().String()
}
