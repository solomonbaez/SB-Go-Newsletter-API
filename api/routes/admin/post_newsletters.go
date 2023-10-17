package routes

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/clients"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/handlers"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/idempotency"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/models"
)

func InsertNewsletter(c *gin.Context, tx pgx.Tx, content *models.Body) (*string, error) {
	id := uuid.NewString()

	query := `INSERT INTO newsletter_issues (
				newsletter_issue_id, 
				title, 
				text_content,
				html_content,
				published_at
			)
			VALUES ($1, $2, $3, $4, now())`
	_, e := tx.Exec(
		c, query, id, content.Title, content.Text, content.Html,
	)
	if e != nil {
		return nil, e
	}

	return &id, e
}

func PostNewsletter(c *gin.Context, dh *handlers.DatabaseHandler, client clients.EmailClient) {
	var newsletter models.Newsletter
	var body models.Body
	var response string
	var e error

	session := sessions.Default(c)
	id := fmt.Sprintf("%v", session.Get("user"))

	requestID := c.GetString("requestID")
	key, _ := c.GetPostForm("idempotency_key")
	session.Set("key", key)
	newsletter.Key = key

	body.Title, _ = c.GetPostForm("title")
	body.Text, _ = c.GetPostForm("text")
	body.Html, _ = c.GetPostForm("html")

	if e = models.ParseNewsletter(&body); e != nil {
		response = "Failed to parse newsletter"
		handlers.HandleError(c, requestID, e, response, http.StatusBadRequest)
		return
	}
	newsletter.Content = &body

	transaction, e := idempotency.TryProcessing(c, dh)
	if e != nil {
		response = "Failed to process transaction"
		handlers.HandleError(c, requestID, e, response, http.StatusInternalServerError)
	}

	if transaction.StartProcessing != nil {
		log.Info().
			Str("requestID", requestID).
			Str("id", id).
			Msg("No saved response, processing request...")

		issue_id, e := InsertNewsletter(c, transaction.StartProcessing, newsletter.Content)
		if e != nil {
			response = "Failed to store newsletter"
			handlers.HandleError(c, requestID, e, response, http.StatusInternalServerError)
			return
		}

		if e := idempotency.EnqueDeliveryTasks(c, transaction.StartProcessing, *issue_id); e != nil {
			response = "Failed to enqueue delivery tasks"
			handlers.HandleError(c, requestID, e, response, http.StatusInternalServerError)
			return
		}

		httpResponse, e := SeeOther(c, "/admin/dashboard")
		if e != nil {
			response = "Failed to parse request body"
			handlers.HandleError(c, requestID, e, response, http.StatusInternalServerError)
			return
		}

		if e := idempotency.SaveResponse(c, dh, httpResponse); e != nil {
			response = "Failed to save http response"
			handlers.HandleError(c, requestID, e, response, http.StatusInternalServerError)
			return
		}

		c.Redirect(http.StatusSeeOther, "dashboard")
	} else {
		log.Info().
			Str("requestID", requestID).
			Str("id", id).
			Msg("Fetched saved response")

		httpResponse := transaction.SavedResponse
		status := httpResponse.StatusCode
		headers := httpResponse.Header

		if status == http.StatusSeeOther {
			location := headers.Get("Location")
			c.Redirect(status, location)

		} else {
			c.JSON(status, headers)
		}
	}

}

func SeeOther(c *gin.Context, location string) (response *http.Response, e error) {
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

	response.Header.Set("Location", location)

	responseBytes, e := io.ReadAll(c.Request.Body)
	if e != nil {
		return nil, e
	}
	response.Body = io.NopCloser(bytes.NewReader(responseBytes))

	return response, e
}

// may need this logic
// subscribers := GetConfirmedSubscribers(c, dh)
// for _, s := range subscribers {

// 	log.Info().
// 		Str("requestID", requestID).
// 		Str("subscriber", s.Email.String()).
// 		Msg("Email sent")
// }
