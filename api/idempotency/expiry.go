package idempotency

import (
	"context"
	"fmt"
	"time"

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
		err = fmt.Errorf("failed to delete expired idempotency key: %w", e)
		log.Error().
			Err(err).
			Msg("")

		return
	}

	query = "DELETE FROM idempotency_headers WHERE idempotency_key = $1"
	_, e = tx.Exec(c, query, key)
	if e != nil {
		err = fmt.Errorf("failed to delete expired idempotency headers: %w", e)
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

// Implemented as a general purpose database sweep
func PruneIdempotencyKeys(c context.Context, tx pgx.Tx, expiration time.Time) (err error) {
	query := "DELETE FROM idempotency WHERE created <= $1"
	_, e := tx.Exec(c, query, expiration)
	if e != nil {
		err = fmt.Errorf("failed to prune expired idempotency keys: %w", e)
		log.Error().
			Err(err).
			Msg("")

		return
	}

	query = "DELETE FROM idempotency_headers WHERE created <= $1"
	_, e = tx.Exec(c, query, expiration)
	if e != nil {
		err = fmt.Errorf("failed to prune expired idempotency headers: %w", e)
		log.Error().
			Err(err).
			Msg("")

		return
	}

	log.Info().
		Msg("Successfully pruned idempotency keys")

	return
}
