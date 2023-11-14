package idempotency

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"

	"github.com/solomonbaez/SB-Go-Newsletter-API/api/handlers"
)

type NextAction struct {
	StartProcessing pgx.Tx
	SavedResponse   *http.Response
}

// TODO implement idempotency expiry
func TryProcessing(c context.Context, dh *handlers.DatabaseHandler, id, key string) (next *NextAction, err error) {
	next = &NextAction{StartProcessing: nil, SavedResponse: nil}

	tx, e := dh.DB.Begin(c)
	defer tx.Rollback(c)
	if e != nil {
		err = fmt.Errorf("failed to begin transaction: %w", e)
		return
	}

	query := "INSERT INTO idempotency (id, idempotency_key, created) VALUES ($1, $2, now()) ON CONFLICT DO NOTHING"
	idempotencyRows, e := tx.Exec(c, query, id, key)
	if e != nil {
		e = fmt.Errorf("failed to insert idempotency log: %w", e)
		log.Error().
			Err(e).
			Msg("idempotency log failure")

		err = e
		return
	}

	query = "INSERT INTO idempotency_headers (idempotency_key, created) VALUES ($1, now())"
	headerRows, e := tx.Exec(c, query, key)
	if e != nil {
		e = fmt.Errorf("failed to insert idempotency header log: %w", e)
		log.Error().
			Err(e).
			Msg("idempotency log failure")

		err = e
		return
	}

	if idempotencyRows.RowsAffected() > 0 && headerRows.RowsAffected() > 0 {
		next.StartProcessing = tx
		return
	}

	savedResponse, e := GetSavedResponse(c, dh, id, key)
	if e != nil {
		e = fmt.Errorf("failed to save response: %w", e)
		log.Error().
			Err(e).
			Msg("failed to save response")

		err = e
		return
	}

	next.SavedResponse = savedResponse
	return
}
