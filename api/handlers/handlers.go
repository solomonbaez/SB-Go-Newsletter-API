package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rs/zerolog/log"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/email"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/models"
)

type Database interface {
	Exec(c context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(c context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(c context.Context, sql string, args ...interface{}) pgx.Row
}

type RouteHandler struct {
	DB Database
}

func NewRouteHandler(db Database) *RouteHandler {
	return &RouteHandler{
		DB: db,
	}
}

type Loader struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

func (rh RouteHandler) Subscribe(c *gin.Context, client email.EmailClient) {
	var subscriber models.Subscriber
	var loader Loader

	requestID := c.GetString("requestID")

	id := uuid.NewString()
	created := time.Now()

	if e := c.ShouldBindJSON(&loader); e != nil {
		response := "Could not subscribe"
		HandleError(c, requestID, e, response, http.StatusBadRequest)
		return
	}

	log.Info().
		Str("requestID", requestID).
		Msg("Validating inputs...")

	email, e := models.ParseEmail(loader.Email)
	if e != nil {
		response := "Could not subscribe"
		HandleError(c, requestID, e, response, http.StatusBadRequest)
		return
	}
	name, e := models.ParseName(loader.Name)
	if e != nil {
		response := "Could not subscribe"
		HandleError(c, requestID, e, response, http.StatusBadRequest)
		return
	}

	subscriber = models.Subscriber{
		Email: email,
		Name:  name,
	}

	// correlate request with inputs
	log.Info().
		Str("requestID", requestID).
		Str("email", subscriber.Email.String()).
		Str("name", subscriber.Name.String()).
		Msg("")

	log.Info().
		Str("requestID", requestID).
		Msg("Subscribing...")

	query := "INSERT INTO subscriptions (id, email, name, created) VALUES ($1, $2, $3, $4)"
	_, e = rh.DB.Exec(c, query, id, subscriber.Email.String(), subscriber.Name.String(), created)
	if e != nil {
		response := "Failed to subscribe"
		HandleError(c, requestID, e, response, http.StatusInternalServerError)
		return
	}

	log.Info().
		Str("requestID", requestID).
		Str("email", subscriber.Email.String()).
		Msg(fmt.Sprintf("Success, %v subscribed!", subscriber.Email.String()))

	c.JSON(http.StatusCreated, gin.H{"requestID": requestID, "subscriber": subscriber})
}

func (rh RouteHandler) GetSubscribers(c *gin.Context) {
	var subscribers []models.Subscriber
	requestID := c.GetString("requestID")

	log.Info().
		Str("requestID", requestID).
		Msg("Fetching subscribers...")

	rows, e := rh.DB.Query(c, "SELECT * FROM subscriptions")
	if e != nil {
		response := "Failed to fetch subscribers"
		HandleError(c, requestID, e, response, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	subscribers, e = pgx.CollectRows[models.Subscriber](rows, BuildSubscriber)
	if e != nil {
		response := "Failed to parse subscribers"
		HandleError(c, requestID, e, response, http.StatusInternalServerError)
		return
	}

	if len(subscribers) > 0 {
		c.JSON(http.StatusOK, gin.H{"requestID": requestID, "subscribers": subscribers})
	} else {
		response := "No subscribers"
		log.Info().
			Str("requestID", requestID).
			Msg(response)

		c.JSON(http.StatusOK, gin.H{"requestID": requestID, "subscribers": response})
	}
}

func (rh RouteHandler) GetSubscriberByID(c *gin.Context) {
	requestID := c.GetString("requestID")

	log.Info().
		Str("requestID", requestID).
		Msg("Validating ID...")

	// Validate UUID
	u := c.Param("id")
	id, e := uuid.Parse(u)
	if e != nil {
		response := "Invalid ID format"
		HandleError(c, requestID, e, response, http.StatusBadRequest)
		return
	}

	log.Info().
		Str("requestID", requestID).
		Msg("Fetching subscriber...")

	var subscriber models.Subscriber
	e = rh.DB.QueryRow(c, "SELECT id, email, name FROM subscriptions WHERE id=$1", id).Scan(&subscriber.ID, &subscriber.Email, &subscriber.Name)
	if e != nil {
		var response string
		if e == pgx.ErrNoRows {
			response = "Subscriber not found"
		} else {
			response = "Database query error"
		}
		HandleError(c, requestID, e, response, http.StatusNotFound)
		return
	}

	c.JSON(http.StatusFound, gin.H{"requestID": requestID, "subscriber": subscriber})
}

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, "OK")
}

func BuildSubscriber(row pgx.CollectableRow) (models.Subscriber, error) {
	var id string
	var email models.SubscriberEmail
	var name models.SubscriberName
	var created time.Time

	e := row.Scan(&id, &email, &name, &created)
	s := models.Subscriber{
		ID:    id,
		Email: email,
		Name:  name,
	}

	return s, e
}

func HandleError(c *gin.Context, id string, e error, response string, status int) {
	log.Error().
		Str("requestID", id).
		Err(e).
		Msg(response)

	c.JSON(status, gin.H{"requestID": id, "error": response + ": " + e.Error()})
}
