package idempotency

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"

	"github.com/solomonbaez/SB-Go-Newsletter-API/api/handlers"
)

type NextAction struct {
	StartProcessing pgx.Tx
	SavedResponse   *http.Response
}

func TryProcessing(c context.Context, dh *handlers.DatabaseHandler, id, key string) (next *NextAction, e error) {
	nextAction := &NextAction{StartProcessing: nil, SavedResponse: nil}

	tx, e := dh.DB.Begin(c)
	if e != nil {
		return nextAction, e
	}

	query := "INSERT INTO idempotency (id, idempotency_key, created) VALUES ($1, $2, now()) ON CONFLICT DO NOTHING"
	idempotencyRows, e := tx.Exec(c, query, id, key)
	if e != nil {
		log.Error().
			Err(e).
			Msg("idempotency")
		return nextAction, e
	}

	query = "INSERT INTO idempotency_headers (idempotency_key) VALUES ($1)"
	headerRows, e := tx.Exec(c, query, key)
	if e != nil {
		log.Error().
			Err(e).
			Msg("headers")
		return nextAction, e
	}

	if idempotencyRows.RowsAffected() > 0 && headerRows.RowsAffected() > 0 {
		nextAction.StartProcessing = tx
		return nextAction, nil
	}

	savedResponse, e := GetSavedResponse(c, dh, id, key)
	if e != nil {
		log.Error().
			Err(e).
			Msg("Saved Response")
		return nil, e
	}

	nextAction.SavedResponse = savedResponse
	return nextAction, nil
}
