package workers

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/clients"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/handlers"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/models"
)

type Task struct {
	NewsletterIssueID string
	SubscriberEmail   models.SubscriberEmail
}

// TODO implement n_retries + execute_after columns to issue_delivery_queue to attempt retries
func TryExecuteTask(c context.Context, dh *handlers.DatabaseHandler, client *clients.SMTPClient) ExecutionOutcome {
	resultChan := make(chan ExecutionOutcome)
	go func() {
		defer close(resultChan)

		task, tx, e := DequeTask(c, dh)
		if e != nil {
			resultChan <- ExecutionOutcomeEmptyQueue
			return
		}

		// re-parse email to ensure data integrity
		var newsletter models.Newsletter
		newsletter.Recipient, e = models.ParseEmail(task.SubscriberEmail.String())
		if e != nil {
			resultChan <- ExecutionOutcomeError
			return
		}

		// TODO add confirmation email logic
		newsletter.Content, e = GetIssue(c, tx, task.NewsletterIssueID)
		if e != nil {
			resultChan <- ExecutionOutcomeError
			return
		}

		if e = models.ParseNewsletter(&newsletter); e != nil {
			resultChan <- ExecutionOutcomeError
			return
		}
		if e = client.SendEmail(&newsletter); e != nil {
			resultChan <- ExecutionOutcomeError
			return
		}

		if e = DeleteTask(c, tx, task); e != nil {
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

func DequeTask(c context.Context, dh *handlers.DatabaseHandler) (task *Task, tx pgx.Tx, err error) {
	var e error
	tx, e = dh.DB.Begin(c)
	if e != nil {
		err = fmt.Errorf("failed to begin transaction: %w", e)
		return
	}

	task = &Task{}
	query := `SELECT newsletter_issue_id, subscriber_email
			FROM issue_delivery_queue
			FOR UPDATE
			SKIP LOCKED
			LIMIT 1`
	e = tx.QueryRow(c, query).Scan(&task.NewsletterIssueID, &task.SubscriberEmail)
	if e != nil {
		err = fmt.Errorf("failed to deque delivery task: %w", e)
		return
	}

	return
}

func GetIssue(c context.Context, tx pgx.Tx, issueID string) (content *models.Body, err error) {
	content = &models.Body{}
	query := `SELECT title, text_content, html_content
			FROM newsletter_issues
			WHERE newsletter_issue_id = $1`
	if e := tx.QueryRow(c, query, issueID).Scan(&content.Title, &content.Text, &content.Html); e != nil {
		err = fmt.Errorf("failed to retrieve newsletter issue: %w", e)
		return
	}

	return
}

func DeleteTask(c context.Context, tx pgx.Tx, task *Task) (err error) {
	query := `DELETE FROM issue_delivery_queue
			WHERE 
			newsletter_issue_id = $1 AND
			subscriber_email = $2`
	_, e := tx.Exec(c, query, task.NewsletterIssueID, task.SubscriberEmail.String())
	if e != nil {
		err = fmt.Errorf("failed to delete delivery task")
		return
	}

	tx.Commit(c)
	return
}

func EnqueDeliveryTasks(c context.Context, tx pgx.Tx, newsletterIssueId string) (err error) {
	query := `INSERT INTO issue_delivery_queue (
				newsletter_issue_id,
				subscriber_email
			)
			SELECT $1, email
			FROM subscriptions
			WHERE status = 'confirmed'`
	_, e := tx.Exec(c, query, newsletterIssueId)
	if e != nil {
		err = fmt.Errorf("failed to enque delivery task")
		return
	}

	tx.Commit(c)
	return
}

// TODO expand confirmation task logic -> new worker pool or mixed concerns?
func EnqueConfirmationTasks(c context.Context, tx pgx.Tx, subscriberEmail string) (err error) {
	// base confirmation email == 0
	query := `INSERT INTO issue_delivery_queue (
				newsletter_issue_id,
				subscriber_email
			)
			VALUES ($1, $2)`
	_, e := tx.Exec(c, query, "0", subscriberEmail)
	if e != nil {
		err = fmt.Errorf("failed to enque confirmation task: %w", e)
		return
	}

	tx.Commit(c)
	return
}
