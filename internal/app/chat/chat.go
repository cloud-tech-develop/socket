package chat

import (
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

func HandleWebSocket(c echo.Context) error {
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
		ID:         clientID,
		Connection: conn,
	}

	if err := joinRoom(client, roomID); err != nil {
		log.Printf("Error al unirse a la sala: %s", err)
		conn.Close()
		return echo.NewHTTPError(500, "Error al unirse a la sala")
	}

	defer func() {
		leaveRoom(client)
		conn.Close()
	}()

	log.Printf("Cliente %s conectado a la sala %s\n", client.ID, roomID)

	for {
		var message Message
		err := conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Cliente %s desconectado de la sala '%s'\n", client.ID, roomID)
				break
			}
			log.Printf("Error al leer el mensaje: %s\n", err)
			break
		}
		log.Printf("Mensaje recibido de %s en la sala '%s': %+v\n", client.ID, roomID, message)
		broadcast(client.Room, message)
	}

	return nil
}

func joinRoom(client *Client, roomID string) error {
	mu.Lock()
	defer mu.Unlock()

	if _, ok := rooms[roomID]; !ok {
		rooms[roomID] = &Room{
			ID:      roomID,
			Clients: make(map[*Client]bool),
		}
	}

	client.Room = rooms[roomID]
	client.Room.Clients[client] = true
	return nil
}

func leaveRoom(client *Client) {
	mu.Lock()
	defer mu.Unlock()

	if client.Room != nil {
		delete(client.Room.Clients, client)

		if len(client.Room.Clients) == 0 {
			delete(rooms, client.Room.ID)
		}
	}
}

func broadcast(room *Room, message Message) {
	message.DateTime = time.Now()
	room.mu.Lock()
	defer room.mu.Unlock()

	for client := range room.Clients {
		err := client.Connection.WriteJSON(message)
		if err != nil {
			log.Printf("Error al enviar mensaje a %s: %s\n", client.ID, err)
			handleClientError(client, err)
		}
	}
}

func handleClientError(client *Client, err error) {
	if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
		log.Printf("Cierre inesperado de la conexión de %s: %s", client.ID, err)
	} else {
		log.Printf("Error de comunicación con el cliente %s: %s", client.ID, err)
	}

	client.Connection.Close()
	delete(client.Room.Clients, client)
}

func generateUniqueID() string {
	return uuid.New().String()
}
