package idempotency

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5"

	"github.com/solomonbaez/SB-Go-Newsletter-API/api/handlers"
)

type NextAction struct {
	StartProcessing pgx.Tx
	SavedResponse   *http.Response
}

func TryProcessing(c context.Context, dh *handlers.DatabaseHandler, id, key string) (*NextAction, error) {
	var query string
	var e error

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
