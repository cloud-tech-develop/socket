package chat

import (
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func (h *Hub) HandleRestRequest(c echo.Context) error {
	var mensaje Message
	if err := c.Bind(&mensaje); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Formato de mensaje JSON inválido"})
	}
	mensaje.DateTime = time.Now()

	roomID := c.QueryParam("room")
	if roomID == "" {
		roomID = "public"
	}

	log.Printf("REST request received for room %s: %+v\n", roomID, mensaje)

	// Envía el mensaje a todos los clientes en la sala usando el método Broadcast del Hub
	h.Broadcast(roomID, mensaje)

	return c.JSON(http.StatusOK, map[string]string{"message": "Mensaje enviado al socket con éxito"})
}
