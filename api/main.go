package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	router.GET("/health", HealthCheck)
	router.Run(":8000")
}

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, "OK")
}
