package utils

import "os"

func GetPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8889"
	}
	return ":" + port
}
