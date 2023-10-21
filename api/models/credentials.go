package models

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"

	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
)

const InvalidRunes = "{}/\\<>() "

type Credentials struct {
	Username string
	Password string
}

type hashParams struct {
	saltLen    uint32
	iterations uint32
	memory     uint32
	threads    uint8
	keyLen     uint32
}

var params = hashParams{
	saltLen:    16,
	iterations: 1,
	memory:     64 * 1024,
	threads:    4,
	keyLen:     32,
}

var BaseHash string

func init() {
	randomKey := uuid.NewString()
	BaseHash, _ = GeneratePHC(randomKey)
}

func ValidatePHC(password string, phc string) (err error) {
	p, s, h, e := DecodePHC(phc)
	if e != nil {
		err = fmt.Errorf("failed to decode PHC: %w", e)
		return
	}

	k := argon2.IDKey([]byte(password), s, p.iterations, p.memory, p.threads, p.keyLen)
	// ctc to protect against timing attacks
	if subtle.ConstantTimeCompare(h, k) != 1 {
		err = errors.New("PHC are not equivalent")
		return
	}

	return
}

func GeneratePHC(password string) (phc string, err error) {
	salt, e := GenerateSalt(params.saltLen)
	if e != nil {
		err = fmt.Errorf("failed to generate PHC salt: %w", e)
		return
	}

	hash := argon2.IDKey([]byte(password), salt, params.iterations, params.memory, params.threads, params.keyLen)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	phc = fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, params.memory, params.iterations, params.threads, b64Salt, b64Hash,
	)

	return
}

func DecodePHC(phc string) (p *hashParams, s, h []byte, err error) {
	values := strings.Split(phc, "$")
	if len(values) != 6 {
		err = errors.New("invalid PHC format")
		return
	}

	var version int
	_, e := fmt.Sscanf(values[2], "v=%d", &version)
	if e != nil {
		err = fmt.Errorf("invalid PHC version: %w", e)
		return
	}
	if version != argon2.Version {
		err = errors.New("incorrect PHC version")
		return
	}

	p = &hashParams{}
	_, e = fmt.Sscanf(values[3], "m=%d,t=%d,p=%d", &p.memory, &p.iterations, &p.threads)
	if e != nil {
		err = errors.New("invalid PHC hash parameters")
		return
	}

	s, e = base64.RawStdEncoding.Strict().DecodeString(values[4])
	if e != nil {
		err = fmt.Errorf("failed to decode PHC salt: %w", e)
		return
	}
	p.saltLen = uint32(len(s))
	h, e = base64.RawStdEncoding.Strict().DecodeString(values[5])
	if e != nil {
		err = fmt.Errorf("failed to decode PHC hash: %w", e)
		return
	}
	p.keyLen = uint32(len(h))

	return
}

func GenerateSalt(s uint32) (b []byte, err error) {
	b = make([]byte, s)
	_, e := rand.Read(b)
	if e != nil {
		err = fmt.Errorf("failed to generate salt: %w", e)
		return
	}

	return b, nil
}
