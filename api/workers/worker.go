package workers

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/clients"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/handlers"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/idempotency"
)

type ExecutionOutcome int

const (
	ExecutionOutcomeEmptyQueue ExecutionOutcome = iota
	ExecutionOutcomeError
	ExecutionOutcomeTaskCompleted
)

func WorkerLoop(c context.Context, dh *handlers.DatabaseHandler, client *clients.SMTPClient) {
	resultChan := make(chan ExecutionOutcome)

	go func() {
		for {
			resultChan <- TryExecuteTask(c, dh, client)
		}
	}()
	for outcome := range resultChan {
		switch outcome {
		case ExecutionOutcomeEmptyQueue:
			log.Info().
				Msg("Empty queue")
			time.Sleep(10 * time.Second)
		case ExecutionOutcomeError:
			log.Error().
				Msg("Failed to complete task")
			time.Sleep(1 * time.Second)
		case ExecutionOutcomeTaskCompleted:
			log.Info().
				Msg("Task complete")
		}
	}
}

func WorkerLoopWithPruning(c context.Context, dh *handlers.DatabaseHandler, client *clients.SMTPClient) {
	resultChan := make(chan ExecutionOutcome)
	keyCleanupTimer := time.NewTimer(24 * time.Hour)

	for {
		select {
		case <-keyCleanupTimer.C:
			tryKeyCleanup(c, dh)
			keyCleanupTimer.Reset(24 * time.Hour)

		case resultChan <- TryExecuteTask(c, dh, client):
			outcome := <-resultChan
			switch outcome {
			case ExecutionOutcomeEmptyQueue:
				log.Info().
					Msg("Empty queue")
				time.Sleep(10 * time.Second)
			case ExecutionOutcomeError:
				log.Error().
					Msg("Failed to complete task")
				time.Sleep(1 * time.Second)
			case ExecutionOutcomeTaskCompleted:
				log.Info().
					Msg("Task complete")
			}
		}
	}
}

func tryKeyCleanup(c context.Context, dh *handlers.DatabaseHandler) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			expiration := time.Now().Add(-24 * time.Hour)

			_, cancel := context.WithTimeout(c, 10*time.Second)
			defer cancel()

			tx, e := dh.DB.Begin(c)
			if e != nil {
				log.Error().
					Err(e).
					Msg("")

				cancel()
			}

			if e := idempotency.PruneIdempotencyKeys(c, tx, expiration); e != nil {
				log.Error().
					Err(e).
					Msg("")

				cancel()
			}

			tx.Commit(c)
		}
	}
}
