package workers

import (
	"context"
	"fmt"
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

func DeliveryWorker(c context.Context, dh *handlers.DatabaseHandler, client *clients.SMTPClient) {
	resultChan := make(chan ExecutionOutcome)

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-c.Done():
				log.Info().
					Msg("worker exit")
				return
			case <-ticker.C:
				resultChan <- TryExecuteTask(c, dh, client)
			}
		}
	}()

	for outcome := range resultChan {
		switch outcome {
		case ExecutionOutcomeEmptyQueue:
			log.Info().
				Msg("Empty queue" + time.Now().String())
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

const pruningInterval = 12
const timeoutInterval = 10

func PruningWorker(c context.Context, dh *handlers.DatabaseHandler) {
	ticker := time.NewTicker(pruningInterval * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			expiration := time.Now().Add(-1 * pruningInterval * time.Hour)
			_, cancel := context.WithTimeout(c, timeoutInterval*time.Second)
			defer cancel()

			if e := idempotency.PruneIdempotencyKeys(c, dh, expiration); e != nil {
				err := fmt.Errorf("failed to prune expired idempotency keys: %w", e)
				log.Error().
					Err(err).
					Msg("")

				continue
			}
		case <-c.Done():
			return
		}
	}
}
