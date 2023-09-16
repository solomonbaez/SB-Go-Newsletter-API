package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/models"
)

const (
	MaxEmailLen = 100
	MaxNameLen  = 100
)

var subscribers = make(map[string]models.Subscriber)

func Subscribe(c *gin.Context) {
	var subscriber models.Subscriber

	if e := c.ShouldBindJSON(&subscriber); e != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not subscribe: " + e.Error()})
		return
	}

	if _, e := subscribers[subscriber.Email]; e {
		c.JSON(http.StatusFound, gin.H{"error": "Email already associated with a subscriber"})
		return

	} else if len(subscriber.Email) > MaxEmailLen {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email exceeds the maximum limit of 100 characters"})
		return

	} else if len(subscriber.Name) > MaxNameLen {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name exceeds the maximum length of 100 characters"})
		return
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
