package blog

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"

	"github.com/solomonbaez/hyacinth/api/handlers"
	"github.com/solomonbaez/hyacinth/api/models"
)

func GetNewlsetterIssues(c *gin.Context, dh *handlers.DatabaseHandler) {
	var newsletterIssues []*models.Newsletter

	requestID := c.GetString("requestID")

	log.Info().
		Str("requestID", requestID).
		Msg("Fetching newsletter issues...")

	var response string
	rows, e := dh.DB.Query(c, "SELECT title, text_content, html_content FROM newsletter_issues")
	if e != nil {
		response = "Failed to fetch newsletter issues"
		handlers.HandleError(c, requestID, e, response, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	newsletterIssues, e = pgx.CollectRows[*models.Newsletter](rows, buildNewsletter)
	if e != nil {
		response = "Failed to parse newsletters"
		handlers.HandleError(c, requestID, e, response, http.StatusInternalServerError)
		return
	}

	if len(newsletterIssues) > 0 {
		response = "No newsletter issues"
		log.Info().
			Str("requestID", requestID).
			Msg(response)
	}

	c.JSON(http.StatusOK, gin.H{"requestID": requestID, "newsletterIssues": newsletterIssues})
}

func GetNewlsetterIssueByTitle(c *gin.Context, dh *handlers.DatabaseHandler) {
	var newsletter = models.Newsletter{}
	newsletter.Content = &models.Body{}

	requestID := c.GetString("requestID")
	title := c.Param("title")

	log.Info().
		Str("requestID", requestID).
		Msg("Fetching newsletter issues...")

	var response string
	e := dh.DB.QueryRow(c, "SELECT title, text_content, html_content FROM newsletter_issues WHERE title=$1", title).
		Scan(&newsletter.Content.Title, &newsletter.Content.Text, &newsletter.Content.Html)
	if e != nil {
		response = "Failed to fetch newsletter"
		handlers.HandleError(c, requestID, e, response, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"requestID": requestID, "newsletter": newsletter})
}

func buildNewsletter(row pgx.CollectableRow) (newsletter *models.Newsletter, err error) {
	var title string
	var text string
	var html string

	if e := row.Scan(&title, &text, &html); e != nil {
		err = fmt.Errorf("database error: %w", e)
		return
	}

	newsletter = &models.Newsletter{}
	newsletter.Content = &models.Body{
		Title: title,
		Text:  text,
		Html:  html,
	}

	return
}
