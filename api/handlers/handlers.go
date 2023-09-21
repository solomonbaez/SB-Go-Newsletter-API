package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

var (
	EmailRegex = regexp.MustCompile((`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`))
)

func (rh RouteHandler) Subscribe(c *gin.Context) {
	var subscriber models.Subscriber

	// TESTING
	id := uuid.NewString()
	created := time.Now()

	if e := c.ShouldBindJSON(&subscriber); e != nil {
		response := "Could not subscribe"
		log.Error().
			Err(e).
			Msg(response)

		c.JSON(http.StatusBadRequest, gin.H{"error": response + ", " + e.Error()})
		return
	}

	if e := ValidateInputs(subscriber); e != nil {
		log.Error().
			Err(e).
			Msg(e.Error())

		c.JSON(http.StatusBadRequest, gin.H{"error": e.Error()})
		return
	}

	query := "INSERT INTO subscriptions (id, email, name, created) VALUES ($1, $2, $3, $4)"
	_, e := rh.DB.Exec(c, query, id, subscriber.Email, subscriber.Name, created)
	if e != nil {
		response := "Failed to subscribe"
		log.Error().
			Err(e).
			Msg(response)

		c.JSON(http.StatusInternalServerError, gin.H{"error": response + ", " + e.Error()})
		return
	}

	log.Info().
		Str("email", subscriber.Email).
		Msg(fmt.Sprintf("%v subscribed!", subscriber.Email))

	c.JSON(http.StatusCreated, subscriber)
}

func (rh RouteHandler) GetSubscribers(c *gin.Context) {
	var subscribers []models.Subscriber
	rows, e := rh.DB.Query(c, "SELECT * FROM subscriptions")
	if e != nil {
		response := "Failed to fetch subscribers"
		log.Error().
			Err(e).
			Msg(response)

		c.JSON(http.StatusInternalServerError, gin.H{"error": response + ", " + e.Error()})
		return
	}
	defer rows.Close()

	subscribers, e = pgx.CollectRows[models.Subscriber](rows, BuildSubscriber)
	if e != nil {
		response := "Failed to fetch subscribers"
		log.Error().
			Msg(response)

		c.JSON(http.StatusBadRequest, response)
		return
	}

	if len(subscribers) > 0 {
		c.JSON(http.StatusOK, subscribers)
	} else {
		response := "No subscribers"
		log.Info().
			Msg(response)

		c.JSON(http.StatusOK, response)
	}
}

func (rh RouteHandler) GetSubscriberByEmail(c *gin.Context) {
	email := c.Param("email")
	dummy_subscriber := models.Subscriber{
		Email: email,
		Name:  "valid",
	}
	if e := ValidateInputs(dummy_subscriber); e != nil {
		response := "Failed to validate email"
		log.Error().
			Err(e).
			Msg(response)

		c.JSON(http.StatusBadRequest, gin.H{"error": response + ", " + e.Error()})
	}

	var name string
	e := rh.DB.QueryRow(c, "SELECT name FROM subscriptions WHERE email=$1", email)
	if e != nil {
		response := "Email does not exist"
		log.Error().
			Msg(response)

		c.JSON(http.StatusBadRequest, gin.H{"error": response})
		return
	}

	subscriber := models.Subscriber{
		Email: email,
		Name:  name,
	}

	c.JSON(http.StatusOK, subscriber)
}

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, "OK")
}

func ValidateInputs(s models.Subscriber) error {
	if len(s.Email) > MaxEmailLen {
		return errors.New(
			fmt.Sprintf("Email exceeds maximum length of: %d characters", MaxEmailLen),
		)
	} else if len(s.Name) > MaxNameLen {
		return errors.New(
			fmt.Sprintf("Name exceeds maximum lenght of: %d characters", MaxNameLen),
		)
	} else if !EmailRegex.MatchString(s.Email) {
		return errors.New(
			fmt.Sprintf("Invalid email format"),
		)
	}

	return nil
}

func BuildSubscriber(row pgx.CollectableRow) (models.Subscriber, error) {
	var id string
	var email string
	var name string
	var created time.Time

	e := row.Scan(&id, &email, &name, &created)
	s := models.Subscriber{
		Email: email,
		Name:  name,
	}

	return s, e
}
