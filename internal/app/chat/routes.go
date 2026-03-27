package chat

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func SetUpRoutes(e *echo.Echo) {

	e.GET("/", HandleHello)
	e.GET("/ws", HandleWebSocket)
	e.POST("/rest-sw", Proxy.HandleRestRequest)
}

func HandleHello(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"message": "Hola mundo!"})
}
