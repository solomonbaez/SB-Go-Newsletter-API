package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/clients"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/models"
)

var confirmationLink string
var confirmation = &models.Newsletter{}
var loader *Loader

func (rh *RouteHandler) Subscribe(c *gin.Context, client *clients.SMTPClient) {
	var subscriber models.Subscriber

	requestID := c.GetString("requestID")

	var response string
	var e error

	tx, e := rh.DB.Begin(c)
	if e != nil {
		response = "Failed to begin transaction"
		HandleError(c, requestID, e, response, http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(c)

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

	subscriber = models.Subscriber{
		Email:  subscriberEmail,
		Name:   subscriberName,
		Status: "pending",
	}

	// correlate request with inputs
	log.Info().
		Str("requestID", requestID).
		Str("email", subscriber.Email.String()).
		Str("name", subscriber.Name.String()).
		Msg("Subscribing...")

	if e := insertSubscriber(c, client, tx, subscriber); e != nil {
		response = "Failed to insert subscriber"
		HandleError(c, requestID, e, response, http.StatusInternalServerError)
	}

	log.Info().
		Str("requestID", requestID).
		Str("email", subscriber.Email.String()).
		Msg(fmt.Sprintf("Success, sent a confirmation email to %v", subscriber.Email.String()))

	c.JSON(http.StatusCreated, gin.H{"requestID": requestID, "subscriber": subscriber})
}

func insertSubscriber(c *gin.Context, client *clients.SMTPClient, tx pgx.Tx, subscriber models.Subscriber) error {
	var e error

	newID := uuid.NewString()
	created := time.Now()

	query := "INSERT INTO subscriptions (id, email, name, created, status) VALUES ($1, $2, $3, $4, $5)"
	_, e = tx.Exec(c, query, newID, subscriber.Email.String(), subscriber.Name.String(), created, "pending")
	if e != nil {
		return e
	}

	token, e := generateCSPRNG()
	if e != nil {
		return e
	}

	if client.SmtpServer != "test" {
		confirmation.Recipient = subscriber.Email

		confirmationLink = fmt.Sprintf("%v/%v", baseURL, token)
		confirmation.Content = &models.Body{
			Title: "Please confirm your subscription",
			Text:  fmt.Sprintf("Welcome to our newsletter! Please confirm your subscription at: %v", confirmationLink),
			Html:  fmt.Sprintf("<p>Welcome to our newsletter! Please confirm your subscription at: <a>%v</a></p>", confirmationLink),
		}

		if e := client.SendEmail(confirmation); e != nil {
			return e
		}
	}

	if e := storeToken(c, tx, newID, token); e != nil {
		return e
	}

	return nil
}
