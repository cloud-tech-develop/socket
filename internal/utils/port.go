package utils

import (
	"os"

	"github.com/joho/godotenv"
)

func GetPort() string {
	// Intentar cargar el archivo .env si existe.
	// Si no se encuentra, continuará silenciosamente usando variables de entorno.
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8889"
	}
	return ":" + port
}
