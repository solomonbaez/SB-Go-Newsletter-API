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

func (rh RouteHandler) Subscribe(c *gin.Context) {
	var subscriber models.Subscriber
	var loader Loader

	request_id := c.GetString("request_id")

	id := uuid.NewString()
	created := time.Now()

	if e := c.ShouldBindJSON(&loader); e != nil {
		response := fmt.Sprintf("Could not subscribe, %v", e.Error())
		log.Error().
			Str("request_id", request_id).
			Err(e).
			Msg(response)

		c.JSON(http.StatusBadRequest, gin.H{"request_id": request_id, "error": response})
		return
	}

	log.Info().
		Str("request_id", request_id).
		Msg("Validating inputs...")

	email, e := models.ParseEmail(loader.Email)
	if e != nil {
		response := fmt.Sprintf("Could not subscribe, %v", e.Error())
		log.Error().
			Str("request_id", request_id).
			Err(e).
			Msg(response)

		c.JSON(http.StatusBadRequest, gin.H{"request_id": request_id, "error": response})
		return
	}
	name, e := models.ParseName(loader.Name)
	if e != nil {
		response := fmt.Sprintf("Could not subscribe, %v", e.Error())
		log.Error().
			Str("request_id", request_id).
			Err(e).
			Msg(response)

		c.JSON(http.StatusBadRequest, gin.H{"request_id": request_id, "error": response})
		return
	}

	subscriber = models.Subscriber{
		Email: email,
		Name:  name,
	}

	// correlate request with inputs
	log.Info().
		Str("request_id", request_id).
		Str("email", subscriber.Email.String()).
		Str("name", subscriber.Name.String()).
		Msg("")

	log.Info().
		Str("request_id", request_id).
		Msg("Subscribing...")

	query := "INSERT INTO subscriptions (id, email, name, created) VALUES ($1, $2, $3, $4)"
	_, e = rh.DB.Exec(c, query, id, subscriber.Email.String(), subscriber.Name.String(), created)
	if e != nil {
		response := fmt.Sprintf("Failed to subscribe, %v", e.Error())
		log.Error().
			Str("request_id", request_id).
			Err(e).
			Msg(response)

		c.JSON(http.StatusInternalServerError, gin.H{"request_id": request_id, "error": response})
		return
	}

	log.Info().
		Str("request_id", request_id).
		Str("email", subscriber.Email.String()).
		Msg(fmt.Sprintf("Success, %v subscribed!", subscriber.Email.String()))

	c.JSON(http.StatusCreated, gin.H{"request_id": request_id, "subscriber": subscriber})
}

func (rh RouteHandler) GetSubscribers(c *gin.Context) {
	var subscribers []models.Subscriber
	request_id := c.GetString("request_id")

	log.Info().
		Str("request_id", request_id).
		Msg("Fetching subscribers...")

	rows, e := rh.DB.Query(c, "SELECT * FROM subscriptions")
	if e != nil {
		response := fmt.Sprintf("Failed to fetch subscribers, %v", e.Error())
		log.Error().
			Str("request_id", request_id).
			Err(e).
			Msg(response)

		c.JSON(http.StatusInternalServerError, gin.H{"request_id": request_id, "error": response})
		return
	}
	defer rows.Close()

	subscribers, e = pgx.CollectRows[models.Subscriber](rows, BuildSubscriber)
	if e != nil {
		response := fmt.Sprintf("Failed to fetch subscribers, %v", e.Error())
		log.Error().
			Str("request_id", request_id).
			Msg(response)

		c.JSON(http.StatusBadRequest, gin.H{"request_id": request_id, "error": response})
		return
	}

	if len(subscribers) > 0 {
		c.JSON(http.StatusOK, gin.H{"request_id": request_id, "subscribers": subscribers})
	} else {
		response := "No subscribers"
		log.Info().
			Str("request_id", request_id).
			Msg(response)

		c.JSON(http.StatusOK, gin.H{"request_id": request_id, "subscribers": response})
	}
}

func (rh RouteHandler) GetSubscriberByID(c *gin.Context) {
	request_id := c.GetString("request_id")

	log.Info().
		Str("request_id", request_id).
		Msg("Validating ID...")

	// Validate UUID
	u := c.Param("id")
	id, e := uuid.Parse(u)
	if e != nil {
		response := fmt.Sprintf("Invalid ID format, %v", e.Error())
		log.Error().
			Str("request_id", request_id).
			Err(e).
			Msg(response)

		c.JSON(http.StatusBadRequest, gin.H{"request_id": request_id, "error": response})
		return
	}

	log.Info().
		Str("request_id", request_id).
		Msg("Fetching subscriber...")

	var subscriber models.Subscriber
	e = rh.DB.QueryRow(c, "SELECT id, email, name FROM subscriptions WHERE id=$1", id).Scan(&subscriber.ID, &subscriber.Email, &subscriber.Name)
	if e != nil {
		var response string
		if e == pgx.ErrNoRows {
			response = fmt.Sprintf("Subscriber not found, %v", e.Error())
		} else {
			response = "Database query error"
		}

		log.Error().
			Str("request_id", request_id).
			Err(e).
			Msg(response)

		c.JSON(http.StatusNotFound, gin.H{"request_id": request_id, "error": response})
		return
	}

	c.JSON(http.StatusFound, gin.H{"request_id": request_id, "subscriber": subscriber})
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
