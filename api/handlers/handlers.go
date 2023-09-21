package handlers

import (
	// "fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/models"
)

type RouteHandler struct {
	DB *pgx.Conn
}

func NewRouteHandler(db *pgx.Conn) *RouteHandler {
	return &RouteHandler{
		DB: db,
	}
}

const (
	MaxEmailLen = 100
	MaxNameLen  = 100
)

func (rh RouteHandler) Subscribe(c *gin.Context) {
	return
	// var subscriber models.Subscriber

	// if e := c.ShouldBindJSON(&subscriber); e != nil {
	// 	response := "Could not subscribe"
	// 	log.Error().
	// 		Err(e).
	// 		Msg(response)

	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "Could not subscribe: " + e.Error()})
	// 	return
	// }

	// if _, e := subscribers[subscriber.Email]; e {
	// 	response := "Email already associated with a subscriber"
	// 	log.Error().
	// 		Msg(response)

	// 	c.JSON(http.StatusFound, gin.H{"error": response})
	// 	return

	// } else if len(subscriber.Email) > MaxEmailLen {
	// 	response := "Email exceeds the maximum limit of 100 characters"
	// 	log.Error().
	// 		Msg(response)

	// 	c.JSON(http.StatusBadRequest, gin.H{"error": response})
	// 	return

	// } else if len(subscriber.Name) > MaxNameLen {
	// 	response := "Name exceeds the maximum limit of 100 characters"
	// 	log.Error().
	// 		Msg(response)

	// 	c.JSON(http.StatusBadRequest, gin.H{"error": response})
	// 	return
	// }

	// subscribers[subscriber.Email] = subscriber
	// log.Info().
	// 	Str("email", subscriber.Email).
	// 	Msg(fmt.Sprintf("%v subscribed!", subscriber.Email))

	// c.JSON(http.StatusCreated, subscriber)
}

func (rh RouteHandler) GetSubscribers(c *gin.Context) {
	var subscribers []models.Subscriber
	rows, e := rh.DB.Query(c, "SELECT * FROM subscriptions")
	if e != nil {
		response := "Failed to fetch subscribers"
		log.Error().
			Err(e).
			Msg(response)

		c.JSON(http.StatusInternalServerError, response+" "+e.Error())
		return
	}

	subscribers, e = pgx.CollectRows[models.Subscriber](rows, func(row pgx.CollectableRow) (models.Subscriber, error) {
		var id int
		var email string
		var name string
		var created time.Time

		e := row.Scan(&id, &email, &name, &created)
		s := models.Subscriber{
			Email: email,
			Name:  name,
		}

		return s, e
	})
	if e != nil {
		response := "Failed to fetch subscribers"
		log.Error().
			Msg(response)

		c.JSON(http.StatusBadRequest, response)
	}

	// TESTING
	test_subscriber := models.Subscriber{
		Email: "testing@testing.com",
		Name:  "testing",
	}
	subscribers = append(subscribers, test_subscriber)

	if len(subscribers) > 0 {
		c.JSON(http.StatusOK, subscribers)
	} else {
		response := "No subscribers"
		log.Info().
			Msg(response)

		c.JSON(http.StatusOK, response)
	}
}

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, "OK")
}
