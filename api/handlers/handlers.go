package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/models"
)

var subscribers = make(map[string]models.Subscriber)

func Subscribe(c *gin.Context) {
	var subscriber models.Subscriber
	if e := c.ShouldBindJSON(&subscriber); e != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not subscribe: " + e.Error()})
	}

	subscribers[subscriber.Email] = subscriber
	c.JSON(http.StatusCreated, subscriber)
}

func GetSubscribers(c *gin.Context) {
	c.JSON(http.StatusOK, subscribers)
}

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, "OK")
}
