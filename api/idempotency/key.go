package idempotency

import (
	"fmt"

	"github.com/solomonbaez/SB-Go-Newsletter-API/api/handlers"
)

const keyLen = 25

type IdempotencyKey string

func GenerateIdempotencyKey() (key IdempotencyKey, err error) {
	k, e := handlers.GenerateCSPRNG(keyLen)
	if e != nil {
		err = fmt.Errorf("failed to generate csprng: %w", e)
		return
	}

	key = IdempotencyKey(k)
	return
}

func (key *IdempotencyKey) String() string {
	return string(*key)
}
