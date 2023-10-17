package idempotency

import (
	"fmt"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"

	"github.com/solomonbaez/SB-Go-Newsletter-API/api/clients"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/handlers"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/models"
)

type NextAction struct {
	StartProcessing pgx.Tx
	SavedResponse   *http.Response
}

func TryProcessing(c *gin.Context, dh *handlers.DatabaseHandler) (*NextAction, error) {
	var query string
	var e error

	session := sessions.Default(c)
	id := fmt.Sprintf("%s", session.Get("user"))
	key := fmt.Sprintf("%s", session.Get("key"))

	tx, e := dh.DB.Begin(c)
	if e != nil {
		return nil, e
	}

	query = "INSERT INTO idempotency (id, idempotency_key, created) VALUES ($1, $2, now()) ON CONFLICT DO NOTHING"
	idempotencyRows, e := tx.Exec(c, query, id, key)
	if e != nil {
		return nil, e
	}

	query = "INSERT INTO idempotency_headers (idempotency_key) VALUES ($1)"
	headerRows, e := tx.Exec(c, query, key)
	if e != nil {
		return nil, e
	}

	if idempotencyRows.RowsAffected() > 0 && headerRows.RowsAffected() > 0 {
		return &NextAction{StartProcessing: tx}, nil
	}

	savedResponse, e := GetSavedResponse(c, dh, id, key)
	if e != nil {
		return nil, e
	}

	return &NextAction{SavedResponse: savedResponse}, nil
}

func EnqueDeliveryTasks(c *gin.Context, tx pgx.Tx, newsletterIssueId string) error {
	query := `INSERT INTO issue_delivery_queue (
				newsletter_issue_id,
				subscriber_email
			)
			SELECT $1, email
			FROM subscriptions
			WHERE status = 'confirmed'`
	_, e := tx.Exec(c, query, newsletterIssueId)
	if e != nil {
		return e
	}

	tx.Commit(c)
	return nil
}

// TODO implement n_retries + execute_after columns to issue_delivery_queue to attempt retries
func TryExecuteTask(c *gin.Context, dh *handlers.DatabaseHandler, client *clients.SMTPClient) (e error) {
	issueID, subscriberEmail, tx, e := DequeTask(c, dh)
	if e != nil {
		return e
	}

	content, e := GetIssue(c, tx, *issueID)
	if e != nil {
		return e
	}

	// re-parse email to ensure data integrity
	var newsletter models.Newsletter
	newsletter.Recipient, e = models.ParseEmail(*subscriberEmail)
	if e != nil {
		return e
	}
	newsletter.Content = content
	if e = models.ParseNewsletter(&newsletter); e != nil {
		return e
	}
	if e = client.SendEmail(&newsletter); e != nil {
		return e
	}

	if e = DeleteTask(c, tx, *issueID, *subscriberEmail); e != nil {
		return e
	}

	log.Info().
		Str("subscriber", *subscriberEmail).
		Msg("Email sent")

	return nil
}

func DequeTask(c *gin.Context, dh *handlers.DatabaseHandler) (issueID, subscriberEmail *string, tx pgx.Tx, e error) {
	tx, e = dh.DB.Begin(c)
	if e != nil {
		return nil, nil, nil, e
	}

	query := `SELECT newsletter_issue_id, subscriber_email
			FROM issue_delivery_queue
			FOR UPDATE
			SKIP LOCKED
			LIMIT 1`
	e = tx.QueryRow(c, query).Scan(&issueID, &subscriberEmail)
	if e != nil {
		return nil, nil, nil, e
	}

	return issueID, subscriberEmail, tx, nil
}

func GetIssue(c *gin.Context, tx pgx.Tx, issueID string) (content *models.Body, e error) {
	query := `SELECT title, text_content, html_content
			FROM newsletter_issues
			WHERE newsletter_issue_id = $1`
	if e = tx.QueryRow(c, query, issueID).Scan(&content.Title, &content.Text, &content.Html); e != nil {
		return nil, e
	}

	return content, nil
}

func DeleteTask(c *gin.Context, tx pgx.Tx, issueID, subscriberEmail string) (e error) {
	query := `DELETE FROM issue_delivery_queue
			WHERE 
			newsletter_issue_id = $1 AND
			subscriber_email = $2`
	_, e = tx.Exec(c, query, issueID, subscriberEmail)
	if e != nil {
		return e
	}

	tx.Commit(c)
	return nil
}
