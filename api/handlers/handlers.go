package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/logger"
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
		response := "Email already associated with a subscriber"
		logger.Error(response)

		c.JSON(http.StatusFound, gin.H{"error": response})
		return

	} else if len(subscriber.Email) > MaxEmailLen {
		response := "Email exceeds the maximum limit of 100 characters"
		logger.Error(response)

		c.JSON(http.StatusBadRequest, gin.H{"error": response})
		return

	} else if len(subscriber.Name) > MaxNameLen {
		response := "Name exceeds the maximum limit of 100 characters"
		logger.Error(response)

		c.JSON(http.StatusBadRequest, gin.H{"error": response})
		return
	}

	subscribers[subscriber.Email] = subscriber
	logger.Info(fmt.Sprintf("%v subscribed!", subscriber.Email))

	c.JSON(http.StatusCreated, subscriber)
}

func GetSubscribers(c *gin.Context) {
	if len(subscribers) > 0 {
		c.JSON(http.StatusOK, subscribers)
	} else {
		response := "No subscribers"
		logger.Info(response)

		c.JSON(http.StatusOK, response)
	}
}

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, "OK")
}
