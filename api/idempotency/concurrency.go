package idempotency

import (
	"fmt"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"

	"github.com/solomonbaez/SB-Go-Newsletter-API/api/handlers"
)

type NextAction struct {
	StartProcessing bool
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

	if e := tx.Commit(c); e != nil {
		return nil, e
	}

	if idempotencyRows.RowsAffected() > 0 && headerRows.RowsAffected() > 0 {
		return &NextAction{StartProcessing: true}, nil
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
				subscriber_email,
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
