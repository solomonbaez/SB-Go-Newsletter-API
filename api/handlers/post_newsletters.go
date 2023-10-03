package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/clients"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/models"
)

func (rh *RouteHandler) PostNewsletter(c *gin.Context, client *clients.SMTPClient) {
	var newsletter models.Newsletter
	var body models.Body
	var response string
	var e error

	requestID := c.GetString("requestID")

	if e = c.ShouldBindJSON(&body); e != nil {
		response = "Could not send newsletter"
		HandleError(c, requestID, e, response, http.StatusInternalServerError)
		return
	}
	newsletter.Content = &body

	subscribers := rh.GetConfirmedSubscribers(c)
	for _, s := range subscribers {
		// re-parse email to ensure data integrity
		newsletter.Recipient, e = models.ParseEmail(s.Email.String())
		if e != nil {
			response = fmt.Sprintf("Invalid email: %v", s.Email.String())
			HandleError(c, requestID, e, response, http.StatusConflict)
			continue
		}
		if e = client.SendEmail(c, &newsletter); e != nil {
			response = "Could not send newsletter"
			HandleError(c, requestID, e, response, http.StatusInternalServerError)
			continue
		}
	}

	c.JSON(http.StatusOK, gin.H{"requestID": requestID, "message": "Emails successfully delivered"})
}
