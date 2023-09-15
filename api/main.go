package main

import (
	"github.com/gin-gonic/gin"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/handlers"
)

func main() {
	router := gin.Default()

	router.GET("/health", handlers.HealthCheck)
	router.Run(":8000")
}
