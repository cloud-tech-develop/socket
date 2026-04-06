package monitor

import (
	"socket/pkg/socket_cors"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	upgrader = socket_cors.Upgrader
)

// MonitorHub gestiona múltiples sesiones de monitoreo.
type MonitorHub struct {
	mu       sync.RWMutex
	sessions map[string]*MonitorSession
}

// MonitorSession representa un grupo de monitoreo identificado por un ID.
type MonitorSession struct {
	ID        string
	Providers map[string]*Client // key: ProviderID
	Viewers   map[*Client]bool   // set of Viewers
	State     map[string]any     // key: ProviderID, value: latest data report
	mu        sync.RWMutex
}

// NewMonitorHub crea una nueva instancia de MonitorHub.
func NewMonitorHub() *MonitorHub {
	return &MonitorHub{
		sessions: make(map[string]*MonitorSession),
	}
}

// Client representa un usuario conectado (proveedor o visor).
type Client struct {
	Hub        *MonitorHub
	ID         string
	Role       string // "provider" o "viewer"
	Connection *websocket.Conn
	Session    *MonitorSession
	send       chan MonitorMessage
}

// MonitorMessage define el formato estándar de comunicación.
type MonitorMessage struct {
	Type       string    `json:"type"`       // "connect", "update", "disconnect", "full_state"
	ProviderID string    `json:"providerId"` // ID del proveedor relevante
	Data       any       `json:"data"`       // Datos reportados
	Timestamp  time.Time `json:"timestamp"`
}
