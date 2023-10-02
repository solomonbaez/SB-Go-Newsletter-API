package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/clients"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/models"
)

var confirmationLink string
var confirmation = &models.Newsletter{}
var loader *Loader

func (rh *RouteHandler) Subscribe(c *gin.Context, client *clients.SMTPClient) {
	var subscriber *models.Subscriber

	requestID := c.GetString("requestID")

	newID := uuid.NewString()
	created := time.Now()
	status := "pending"

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
	_, e = tx.Exec(c, query, newID, subscriber.Email.String(), subscriber.Name.String(), created, status)
	if e != nil {
		response = "Failed to subscribe"
		HandleError(c, requestID, e, response, http.StatusInternalServerError)
		return
	}

	token, e := generateCSPRNG()
	if e != nil {
		response = "Failed to generate token"
		HandleError(c, requestID, e, response, http.StatusInternalServerError)
		return
	}

	if client.SmtpServer != "test" {
		confirmationLink = fmt.Sprintf("%v/%v", baseURL, token)
		confirmation.Content.Title = "Please confirm your subscription"
		confirmation.Content.Text = fmt.Sprintf("Welcome to our newsletter! Please confirm your subscription at: %v", confirmationLink)
		confirmation.Content.Html = fmt.Sprintf("<p>Welcome to our newsletter! Please confirm your subscription at: <a>%v</a></p>", confirmationLink)

		confirmation.Recipient = subscriber.Email
		if e := client.SendEmail(c, confirmation); e != nil {
			response = "Failed to send confirmation email"
			HandleError(c, requestID, e, response, http.StatusInternalServerError)
			return
		}
	}

	if e := rh.storeToken(c, tx, newID, token); e != nil {
		response = "Failed to store user token"
		HandleError(c, requestID, e, response, http.StatusInternalServerError)
		return
	}

	log.Info().
		Str("requestID", requestID).
		Str("email", subscriber.Email.String()).
		Msg(fmt.Sprintf("Success, sent a confirmation email to %v", subscriber.Email.String()))

	c.JSON(http.StatusCreated, gin.H{"requestID": requestID, "subscriber": subscriber})
}
