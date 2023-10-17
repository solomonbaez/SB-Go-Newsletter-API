package idempotency

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/clients"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/handlers"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/models"
)

type ExecutionOutcome int

const (
	ExecutionOutcomeEmptyQueue ExecutionOutcome = iota
	ExecutionOutcomeError
	ExecutionOutcomeTaskCompleted
)

func WorkerLoop(c *gin.Context, dh *handlers.DatabaseHandler, client *clients.SMTPClient) {
	resultChan := make(chan ExecutionOutcome)

	go func() {
		for {
			resultChan <- TryExecuteTask(c, dh, client)
		}
	}()
	for outcome := range resultChan {
		switch outcome {
		case ExecutionOutcomeEmptyQueue:
			time.Sleep(10 * time.Second)
		case ExecutionOutcomeError:
			time.Sleep(1 * time.Second)
		case ExecutionOutcomeTaskCompleted:
		}
	}
}

type Task struct {
	NewsletterIssueID string
	SubscriberEmail   models.SubscriberEmail
}

// TODO implement n_retries + execute_after columns to issue_delivery_queue to attempt retries
func TryExecuteTask(c *gin.Context, dh *handlers.DatabaseHandler, client *clients.SMTPClient) ExecutionOutcome {
	resultChan := make(chan ExecutionOutcome)
	go func() {
		defer close(resultChan)

		task, tx, e := DequeTask(c, dh)
		if e != nil {
			log.Error().
				Err(e).
				Msg(e.Error())
			resultChan <- ExecutionOutcomeEmptyQueue
			return
		}

		content, e := GetIssue(c, tx, task.NewsletterIssueID)
		if e != nil {
			log.Error().
				Err(e).
				Msg(e.Error())
			resultChan <- ExecutionOutcomeError
			return
		}

		// re-parse email to ensure data integrity
		var newsletter models.Newsletter
		newsletter.Recipient, e = models.ParseEmail(task.SubscriberEmail.String())
		if e != nil {
			log.Error().
				Err(e).
				Msg(e.Error())
			resultChan <- ExecutionOutcomeError
			return
		}
		newsletter.Content = content
		if e = models.ParseNewsletter(&newsletter); e != nil {
			log.Error().
				Err(e).
				Msg(e.Error())
			resultChan <- ExecutionOutcomeError
			return
		}
		if e = client.SendEmail(&newsletter); e != nil {
			log.Error().
				Err(e).
				Msg(e.Error())
			resultChan <- ExecutionOutcomeError
			return
		}

		if e = DeleteTask(c, tx, task); e != nil {
			log.Error().
				Err(e).
				Msg(e.Error())
			resultChan <- ExecutionOutcomeError
			return
		}

		log.Info().
			Str("subscriber", task.SubscriberEmail.String()).
			Msg("Email sent")

		resultChan <- ExecutionOutcomeTaskCompleted
	}()
	return <-resultChan
}

func DequeTask(c *gin.Context, dh *handlers.DatabaseHandler) (task *Task, tx pgx.Tx, e error) {
	tx, e = dh.DB.Begin(c)
	if e != nil {
		return nil, nil, e
	}

	query := `SELECT newsletter_issue_id, subscriber_email
			FROM issue_delivery_queue
			FOR UPDATE
			SKIP LOCKED
			LIMIT 1`
	e = tx.QueryRow(c, query).Scan(&task.NewsletterIssueID, &task.SubscriberEmail)
	if e != nil {
		return nil, nil, e
	}

	return task, tx, nil
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

func DeleteTask(c *gin.Context, tx pgx.Tx, task *Task) (e error) {
	query := `DELETE FROM issue_delivery_queue
			WHERE 
			newsletter_issue_id = $1 AND
			subscriber_email = $2`
	_, e = tx.Exec(c, query, task.NewsletterIssueID, task.SubscriberEmail.String())
	if e != nil {
		return e
	}

	tx.Commit(c)
	return nil
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
