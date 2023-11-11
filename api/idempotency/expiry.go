package idempotency

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/handlers"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/models"
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
func PruneIdempotencyKeys(c context.Context, dh *handlers.DatabaseHandler, expiration time.Time) (err error) {
	tx, e := dh.DB.Begin(c)
	defer tx.Rollback(c)

	if e != nil {
		err = fmt.Errorf("failed to begin transaction: %w", e)
		log.Error().
			Err(err).
			Msg("failed to begin transaction")

		return
	}

	query := "DELETE FROM idempotency WHERE created <= $1"
	_, e = tx.Exec(c, query, expiration)
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

	if e := tx.Commit(c); e != nil {
		err = fmt.Errorf("failed to commit transaction: %w", e)
		log.Error().
			Err(err).
			Msg("")
	}

	log.Info().
		Msg("Successfully pruned idempotency keys")

	return
}

// TODO implement task within PruningWorker to prune unconfirmed subscribers every 1 week
func PruneUnconfirmedSubscribers(c context.Context, dh *handlers.DatabaseHandler, expiration time.Time) (err error) {
	query := "SELECT (id) FROM subscriptions WHERE created <= $1"
	rows, e := dh.DB.Query(c, query, expiration)
	if e != nil {
		err = fmt.Errorf("failed to fetch expired unconfirmed subscribers: %w", e)
		log.Error().
			Err(err).
			Msg("")

		return
	}

	// TODO: implement limited BuildSubscriber function to only fetch ID
	expiredSubscribers, e := pgx.CollectRows[*models.Subscriber](rows, handlers.BuildSubscriber)
	if e != nil {
		err = fmt.Errorf("failed to fetch expired unconfirmed subscribers: %w", e)
		log.Error().
			Err(err).
			Msg("")

		return
	}

	for _, subscriber := range expiredSubscribers {
		tx, e := dh.DB.Begin(c)
		defer tx.Rollback(c)

		if e != nil {
			err = fmt.Errorf("failed to begin transaction for %s: %w", subscriber.ID, e)
			log.Error().
				Err(err).
				Msg("failed to begin transaction")

			continue
		}

		query = "DELETE FROM subscriptions WHERE id = $1"
		_, e = tx.Exec(c, query, subscriber.ID)
		if e != nil {
			err = fmt.Errorf("failed to delete expired unconfirmed subscriber %s: %w", subscriber.ID, e)
			log.Error().
				Err(err).
				Msg("")

			tx.Rollback(c)
			continue
		}

		query = "DELETE FROM subscription_tokens WHERE subscriber_id = $1"
		_, e = tx.Exec(c, query, subscriber.ID)
		if e != nil {
			err = fmt.Errorf("failed to delete token for expired unconfirmed subscriber %s: %w", subscriber.ID, e)
			log.Error().
				Err(err).
				Msg("")

			tx.Rollback(c)
			continue
		}

		if e := tx.Commit(c); e != nil {
			err = fmt.Errorf("failed to commit transaction for expired unconfirmed subscriber %s: %w", subscriber.ID, e)
			log.Error().
				Err(err).
				Msg("")

			tx.Rollback(c)
			continue
		}
	}

	log.Info().
		Msg("Successfully pruned unconfirmed subscribers")

	return
}
