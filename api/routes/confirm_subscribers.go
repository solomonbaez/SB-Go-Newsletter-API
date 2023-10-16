package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/handlers"
)

func ConfirmSubscriber(c *gin.Context, dh *handlers.DatabaseHandler) {
	var id string
	var query string
	var response string
	var e error

	requestID := c.GetString("requestID")
	token := c.Param("token")

	query = "SELECT subscriber_id FROM subscription_tokens WHERE subscription_token = $1"
	e = dh.DB.QueryRow(c, query, token).Scan(&id)
	if e != nil {
		response = "Failed to fetch subscriber ID"
		handlers.HandleError(c, requestID, e, response, http.StatusInternalServerError)
		return
	}

	query = "UPDATE subscriptions SET status = 'confirmed' WHERE id = $1"
	_, e = dh.DB.Exec(c, query, id)
	if e != nil {
		response = "Failed to confirm subscription"
		handlers.HandleError(c, requestID, e, response, http.StatusInternalServerError)
		return
	}

	log.Info().
		Msg("Subscription confirmed")

	c.JSON(http.StatusAccepted, gin.H{"requestID": requestID, "subscriber": "Subscription confirmed"})
}