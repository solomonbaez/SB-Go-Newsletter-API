package routes

import (
	"fmt"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
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
	// key, _ := c.GetPostForm("idempotency_key")
	key := "test"
	session.Set("key", key)
	newsletter.Key = key

	savedResponse, _ := idempotency.GetSavedResponse(c, dh, id, key)
	// Early return if no response is saved
	if savedResponse != nil {
		log.Info().
			Str("requestID", requestID).
			Str("id", id).
			Msg("Fetched saved response")

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

	httpResponse := SeeOther(c, "/admin/dashboard")

	idempotency.SaveResponse(c, dh, httpResponse)
	c.Redirect(http.StatusSeeOther, "dashboard")
}

func SeeOther(c *gin.Context, location string) (response *http.Response) {
	response = &http.Response{
		Status:        http.StatusText(http.StatusSeeOther),
		StatusCode:    http.StatusSeeOther,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Header:        make(http.Header),
		Request:       c.Request,
		ContentLength: -1, // Set the content length as needed
	}

	// Set the "Location" header
	response.Header.Set("Location", location)

	return response
}