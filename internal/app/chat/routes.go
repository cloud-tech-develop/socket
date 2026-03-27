package chat

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func SetUpRoutes(e *echo.Echo, h *Hub) {
	e.GET("/", HandleHello)
	e.GET("/ws", h.HandleWebSocket)
	e.POST("/rest-sw", h.HandleRestRequest)
}

func HandleHello(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"message": "Hola mundo!"})
}
