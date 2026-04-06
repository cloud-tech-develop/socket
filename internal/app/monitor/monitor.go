package monitor

import (
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 1024
)

// HandleMonitor gestiona la conexión inicial de un cliente.
func (h *MonitorHub) HandleMonitor(c echo.Context) error {
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Printf("Error al actualizar la conexión WebSocket: %s", err)
		return echo.NewHTTPError(400, "Error al actualizar la conexión WebSocket")
	}

	role := c.QueryParam("role")
	if role == "" {
		role = "provider" // Rol por defecto
	}

	clientID := c.QueryParam("id")
	if clientID == "" {
		clientID = uuid.New().String()
	}

	monitorID := c.QueryParam("monitorId")
	if monitorID == "" {
		monitorID = "default"
	}

	client := &Client{
		Hub:        h,
		ID:         clientID,
		Role:       role,
		Connection: conn,
		send:       make(chan MonitorMessage, 256),
	}

	if err := h.joinSession(client, monitorID); err != nil {
		log.Printf("Error al unirse a la sesión: %s", err)
		conn.Close()
		return echo.NewHTTPError(500, "Error al unirse a la sesión de monitoreo")
	}

	log.Printf("Monitor: %s (%s) conectado a la sesión %s\n", client.ID, client.Role, monitorID)

	go client.writePump()
	go client.readPump()

	return nil
}

func (h *MonitorHub) joinSession(client *Client, monitorID string) error {
	h.mu.Lock()
	if _, ok := h.sessions[monitorID]; !ok {
		h.sessions[monitorID] = &MonitorSession{
			ID:        monitorID,
			Providers: make(map[string]*Client),
			Viewers:   make(map[*Client]bool),
			State:     make(map[string]any),
		}
	}
	session := h.sessions[monitorID]
	h.mu.Unlock()

	client.Session = session
	session.mu.Lock()
	defer session.mu.Unlock()

	if client.Role == "provider" {
		// Registrar como proveedor
		session.Providers[client.ID] = client
		// Notificar a los visores
		h.broadcastToViewers(session, MonitorMessage{
			Type:       "connect",
			ProviderID: client.ID,
			Timestamp:  time.Now(),
		})
	} else {
		// Registrar como visor
		session.Viewers[client] = true
		// Enviar estado actual completo al visor nuevo
		client.send <- MonitorMessage{
			Type:      "full_state",
			Data:      session.State,
			Timestamp: time.Now(),
		}
	}

	return nil
}

func (h *MonitorHub) leaveSession(client *Client) {
	session := client.Session
	if session == nil {
		return
	}

	session.mu.Lock()
	if client.Role == "provider" {
		delete(session.Providers, client.ID)
		// No eliminamos del estado inmediatamente si queremos que el visor sepa que estuvo allí
		// pero notificamos la desconexión.
		h.broadcastToViewers(session, MonitorMessage{
			Type:       "disconnect",
			ProviderID: client.ID,
			Timestamp:  time.Now(),
		})
	} else {
		delete(session.Viewers, client)
	}
	
	closeSession := len(session.Providers) == 0 && len(session.Viewers) == 0
	session.mu.Unlock()

	if closeSession {
		h.mu.Lock()
		delete(h.sessions, session.ID)
		h.mu.Unlock()
	}
}

func (h *MonitorHub) broadcastToViewers(session *MonitorSession, message MonitorMessage) {
	// session.mu ya debería estar bloqueado por quien llama.
	for viewer := range session.Viewers {
		select {
		case viewer.send <- message:
		default:
			close(viewer.send)
			delete(session.Viewers, viewer)
		}
	}
}

func (c *Client) readPump() {
	defer func() {
		c.Hub.leaveSession(c)
		c.Connection.Close()
	}()

	c.Connection.SetReadLimit(maxMessageSize)
	c.Connection.SetReadDeadline(time.Now().Add(pongWait))
	c.Connection.SetPongHandler(func(string) error { c.Connection.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		var msg MonitorMessage
		err := c.Connection.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		if c.Role == "provider" {
			// Solo los proveedores pueden enviar actualizaciones de datos
			msg.Type = "update"
			msg.ProviderID = c.ID
			msg.Timestamp = time.Now()

			// Actualizar estado persistente en la sesión
			c.Session.mu.Lock()
			c.Session.State[c.ID] = msg.Data
			c.Hub.broadcastToViewers(c.Session, msg)
			c.Session.mu.Unlock()
		}
		// Los visores no envían mensajes actualmente, si lo hacen se ignoran.
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
