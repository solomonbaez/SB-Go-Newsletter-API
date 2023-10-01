package handlers

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rs/zerolog/log"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/clients"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/models"
)

const confirmationLink = "www.test.com"
const tokenLength = 25
const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// generate new random seed
var seed *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

var confirmation = &clients.Message{
	Subject: "Confirm Your Subscription!",
	Text:    fmt.Sprintf("Welcome to our newsletter, please follow the link to confirm: %v", confirmationLink),
	Html:    fmt.Sprintf("<p>Welcome to our newsletter, please follow the link to confirm: %v</p>", confirmationLink),
}

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

var loader *Loader

func (rh *RouteHandler) Subscribe(c *gin.Context, client *clients.SMTPClient) {
	var subscriber *models.Subscriber

	requestID := c.GetString("requestID")

	newID := uuid.NewString()
	created := time.Now()
	status := "pending"

	var response string
	var e error

	if e = c.ShouldBindJSON(&loader); e != nil {
		response = "Could not subscribe"
		HandleError(c, requestID, e, response, http.StatusBadRequest)
		return
	}

	log.Info().
		Str("requestID", requestID).
		Msg("Validating inputs...")

	subscriberEmail, e := models.ParseEmail(loader.Email)
	if e != nil {
		response = "Could not subscribe"
		HandleError(c, requestID, e, response, http.StatusBadRequest)
		return
	}
	subscriberName, e := models.ParseName(loader.Name)
	if e != nil {
		response := "Could not subscribe"
		HandleError(c, requestID, e, response, http.StatusBadRequest)
		return
	}

	subscriber = &models.Subscriber{
		Email:  subscriberEmail,
		Name:   subscriberName,
		Status: status,
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

	query := "INSERT INTO subscriptions (id, email, name, created, status) VALUES ($1, $2, $3, $4, $5)"
	_, e = rh.DB.Exec(c, query, newID, subscriber.Email.String(), subscriber.Name.String(), created, status)
	if e != nil {
		response = "Failed to subscribe"
		HandleError(c, requestID, e, response, http.StatusInternalServerError)
		return
	}

	token := generateToken()
	if client.SmtpServer != "test" {
		confirmation.Recipient = subscriber.Email
		if e := client.SendEmail(c, confirmation, token); e != nil {
			response = "Failed to send confirmation email"
			HandleError(c, requestID, e, response, http.StatusInternalServerError)
			return
		}
	}

	if e := rh.storeToken(c, newID, token); e != nil {
		response = "Failed to store user token"
		HandleError(c, requestID, e, response, http.StatusInternalServerError)
		return
	}

	log.Info().
		Str("requestID", requestID).
		Str("email", subscriber.Email.String()).
		Msg(fmt.Sprintf("Success, %v subscribed!", subscriber.Email.String()))

	c.JSON(http.StatusCreated, gin.H{"requestID": requestID, "subscriber": subscriber})
}

func (rh *RouteHandler) GetSubscribers(c *gin.Context) {
	var subscribers []*models.Subscriber
	requestID := c.GetString("requestID")

	var response string
	var e error

	log.Info().
		Str("requestID", requestID).
		Msg("Fetching subscribers...")

	rows, e := rh.DB.Query(c, "SELECT * FROM subscriptions")
	if e != nil {
		response = "Failed to fetch subscribers"
		HandleError(c, requestID, e, response, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	subscribers, e = pgx.CollectRows[*models.Subscriber](rows, BuildSubscriber)
	if e != nil {
		response = "Failed to parse subscribers"
		HandleError(c, requestID, e, response, http.StatusInternalServerError)
		return
	}

	if len(subscribers) > 0 {
		c.JSON(http.StatusOK, gin.H{"requestID": requestID, "subscribers": subscribers})
	} else {
		response = "No subscribers"
		log.Info().
			Str("requestID", requestID).
			Msg(response)

		c.JSON(http.StatusOK, gin.H{"requestID": requestID, "subscribers": response})
	}
}

func (rh *RouteHandler) GetSubscriberByID(c *gin.Context) {
	requestID := c.GetString("requestID")

	var response string
	var e error

	log.Info().
		Str("requestID", requestID).
		Msg("Validating ID...")

	// Validate UUID
	u := c.Param("id")
	id, e := uuid.Parse(u)
	if e != nil {
		response = "Invalid ID format"
		HandleError(c, requestID, e, response, http.StatusBadRequest)
		return
	}

	log.Info().
		Str("requestID", requestID).
		Msg("Fetching subscriber...")

	var subscriber models.Subscriber
	e = rh.DB.QueryRow(c, "SELECT id, email, name, status FROM subscriptions WHERE id=$1", id).Scan(&subscriber.ID, &subscriber.Email, &subscriber.Name, &subscriber.Status)
	if e != nil {
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

func (rh *RouteHandler) storeToken(c *gin.Context, id string, token string) error {
	query := "INSERT INTO subscription_tokens (subscription_token, subscriber_id) VALUES ($1, $2)"
	_, e := rh.DB.Exec(c, query, token, id)
	if e != nil {
		return e
	}
	return nil
}

func generateToken() string {
	b := make([]byte, tokenLength)
	for i := range b {
		b[i] = charset[seed.Intn(len(charset))]
	}

	return string(b)
}

func BuildSubscriber(row pgx.CollectableRow) (*models.Subscriber, error) {
	var id string
	var email models.SubscriberEmail
	var name models.SubscriberName
	var created time.Time
	var status string

	e := row.Scan(&id, &email, &name, &created, &status)
	s := &models.Subscriber{
		ID:     id,
		Email:  email,
		Name:   name,
		Status: status,
	}

	return s, e
}

func HandleError(c *gin.Context, id string, e error, response string, status int) {
	log.Error().
		Str("requestID", id).
		Err(e).
		Msg(response)

	var message strings.Builder
	message.WriteString(response)
	message.WriteString(": ")
	message.WriteString(e.Error())

	c.JSON(status, gin.H{"requestID": id, "error": message.String()})
}
