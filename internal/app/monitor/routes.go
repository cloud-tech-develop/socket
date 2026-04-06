package monitor

import (
	"github.com/labstack/echo/v4"
)

// SetUpRoutes registra las rutas del sistema de monitoreo.
func SetUpRoutes(e *echo.Echo, h *MonitorHub) {
	e.GET("/monitor", h.HandleMonitor)
}
