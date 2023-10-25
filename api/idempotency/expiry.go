package idempotency

import (
	"fmt"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

// Implemented as a fallback on request retries
func DeleteIdempotencyKey(c *gin.Context, tx pgx.Tx, key IdempotencyKey) (err error) {
	session := sessions.Default(c)
	id := fmt.Sprintf("%s", session.Get("id"))

	query := "DELETE FROM idempotency WHERE id = $1 AND idempotency_key = $2"
	_, e := tx.Exec(c, query, id, key)
	if e != nil {
		err = fmt.Errorf("failed to delete idempotency key: %w", e)
		log.Error().
			Err(err).
			Msg("")

		return
	}

	query = "DELETE FROM idempotency_headers WHERE idempotency_key = $1"
	_, e = tx.Exec(c, query, key)
	if e != nil {
		err = fmt.Errorf("failed to delete saved idempotency headers: %w", e)
		log.Error().
			Err(err).
			Msg("")

		return
	}

	log.Info().
		Str("id", id).
		Str("key", key.String()).
		Msg("Key expired")

	return
}
