package workers

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/clients"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/handlers"
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
			log.Info().
				Msg("Error")
			time.Sleep(1 * time.Second)
		case ExecutionOutcomeTaskCompleted:
			log.Info().
				Msg("Task complete")
		}
	}
}
