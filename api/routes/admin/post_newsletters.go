package routes

import (
	"fmt"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/clients"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/handlers"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/idempotency"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/models"
)

func PostNewsletter(c *gin.Context, dh *handlers.DatabaseHandler, client clients.EmailClient) {
	var newsletter models.Newsletter
	var body models.Body
	var response string
	var e error

	session := sessions.Default(c)
	id := fmt.Sprintf("%v", session.Get("user"))

	requestID := c.GetString("requestID")
	key, _ := c.GetPostForm("idempotency_key")
	newsletter.Key = key

	savedResponse, _ := idempotency.GetSavedResponse(c, dh, id, key)
	// Early return if no response is saved
	if savedResponse == nil {

		body.Title, _ = c.GetPostForm("title")
		body.Text, _ = c.GetPostForm("text")
		body.Html, _ = c.GetPostForm("html")

		if e = models.ParseNewsletter(&body); e != nil {
			response = "Failed to parse newsletter"
			handlers.HandleError(c, requestID, e, response, http.StatusBadRequest)
			return
		}
		newsletter.Content = &body

		subscribers := GetConfirmedSubscribers(c, dh)
		for _, s := range subscribers {
			// re-parse email to ensure data integrity
			newsletter.Recipient, e = models.ParseEmail(s.Email.String())
			if e != nil {
				response = fmt.Sprintf("Invalid email: %v", s.Email.String())
				handlers.HandleError(c, requestID, e, response, http.StatusConflict)
				continue
			}
			if e = models.ParseNewsletter(&newsletter); e != nil {
				response = "Invalid newsletter"
				handlers.HandleError(c, requestID, e, response, http.StatusBadRequest)
				continue
			}
			if e = client.SendEmail(&newsletter); e != nil {
				response = "Failed to send newsletter"
				handlers.HandleError(c, requestID, e, response, http.StatusInternalServerError)
				continue
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"requestID": requestID, "message": "Emails successfully delivered"})
}
