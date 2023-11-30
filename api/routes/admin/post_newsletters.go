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
	"github.com/solomonbaez/hyacinth/api/clients"
	"github.com/solomonbaez/hyacinth/api/handlers"
	"github.com/solomonbaez/hyacinth/api/idempotency"
	"github.com/solomonbaez/hyacinth/api/models"
	"github.com/solomonbaez/hyacinth/api/workers"
)

func InsertNewsletter(c *gin.Context, tx pgx.Tx, content *models.Body) (id *string, err error) {
	issueID := uuid.NewString()

	query := `INSERT INTO newsletter_issues (
				newsletter_issue_id, 
				title, 
				text_content,
				html_content,
				published_at
			)
			VALUES ($1, $2, $3, $4, now())`
	_, e := tx.Exec(c, query, issueID, content.Title, content.Text, content.Html)
	if e != nil {
		err = fmt.Errorf("failed to insert newsletter issue: %w", e)
		return
	}

	id = &issueID
	return
}

func PostNewsletter(c *gin.Context, dh *handlers.DatabaseHandler, client *clients.SMTPClient) {
	var newsletter models.Newsletter
	var body models.Body

	requestID := c.GetString("requestID")

	session := sessions.Default(c)
	id := fmt.Sprintf("%v", session.Get("id"))
	key, _ := c.GetPostForm("idempotency_key")
	session.Set("key", key)

	body.Title, _ = c.GetPostForm("title")
	body.Text, _ = c.GetPostForm("text")
	body.Html, _ = c.GetPostForm("html")

	var response string
	if e := models.ParseNewsletter(&body); e != nil {
		response = "Failed to parse newsletter"
		handlers.HandleError(c, requestID, e, response, http.StatusBadRequest)
		return
	}
	newsletter.Content = &body

	transaction, e := idempotency.TryProcessing(c, dh, id, key)
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

		if e := workers.EnqueDeliveryTasks(c, transaction.StartProcessing, *issue_id); e != nil {
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

		if e := idempotency.SaveResponse(c, dh, id, key, httpResponse); e != nil {
			response = "Failed to save http response"
			handlers.HandleError(c, requestID, e, response, http.StatusInternalServerError)
			return
		}

		session.AddFlash(fmt.Sprintf("Newsletter %s posted!", *issue_id))
		session.Save()

		c.Header("X-Redirect", "Newsletter")
		c.Redirect(http.StatusSeeOther, "dashboard")
		return

	} else if transaction.SavedResponse != nil {
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
			return
		}
	}

	c.Header("X-Redirect", "Fatal")
	c.Redirect(http.StatusSeeOther, "newsletter")
}

func SeeOther(c *gin.Context, location string) (response *http.Response, err error) {
	response = &http.Response{
		Status:        http.StatusText(http.StatusSeeOther),
		StatusCode:    http.StatusSeeOther,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Header:        make(http.Header),
		Request:       c.Request,
		ContentLength: -1,
	}

	response.Header.Set("Location", location)
	responseBytes, e := io.ReadAll(c.Request.Body)
	if e != nil {
		err = fmt.Errorf("failed to read request body: %w", e)
		return
	}
	response.Body = io.NopCloser(bytes.NewReader(responseBytes))
	return
}
