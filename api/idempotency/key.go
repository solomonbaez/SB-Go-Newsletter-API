package idempotency

import "github.com/solomonbaez/SB-Go-Newsletter-API/api/handlers"

const keyLen = 25

type IdempotencyKey string

func GenerateIdempotencyKey() (*IdempotencyKey, error) {
	key, e := handlers.GenerateCSPRNG(keyLen)
	if e != nil {
		return nil, e
	}

	idempotencyKey := IdempotencyKey(key)
	return &idempotencyKey, nil
}

func (key IdempotencyKey) String() string {
	return string(key)
}
