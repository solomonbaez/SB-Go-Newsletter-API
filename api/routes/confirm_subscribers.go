package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/solomonbaez/hyacinth/api/handlers"
)

func ConfirmSubscriber(c *gin.Context, dh *handlers.DatabaseHandler) {
	var response string

	requestID := c.GetString("requestID")
	token := c.Param("token")

	var id string
	query := "SELECT subscriber_id FROM subscription_tokens WHERE subscription_token = $1"
	e := dh.DB.QueryRow(c, query, token).Scan(&id)
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
		Str("id", id).
		Msg("Subscription confirmed")

	c.JSON(http.StatusAccepted, gin.H{"requestID": requestID, "subscriber": "Subscription confirmed"})
}
