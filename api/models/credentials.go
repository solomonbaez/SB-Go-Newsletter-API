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
	r := uuid.NewString()
	BaseHash, _ = GeneratePHC(r)
}

func ValidatePHC(password string, phc string) error {
	p, s, h, e := DecodePHC(phc)
	if e != nil {
		return e
	}

	k := argon2.IDKey([]byte(password), s, p.iterations, p.memory, p.threads, p.keyLen)

	// ctc to protect against timing attacks
	if subtle.ConstantTimeCompare(h, k) != 1 {
		e = errors.New("PHC are not equivalent")
		return e
	}

	return nil
}

func GeneratePHC(password string) (string, error) {
	salt, e := GenerateSalt(params.saltLen)
	if e != nil {
		return "", e
	}

	hash := argon2.IDKey([]byte(password), salt, params.iterations, params.memory, params.threads, params.keyLen)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encodedHash := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, params.memory, params.iterations, params.threads, b64Salt, b64Hash,
	)

	return encodedHash, nil
}

func DecodePHC(phc string) (p *hashParams, s, h []byte, e error) {
	values := strings.Split(phc, "$")
	if len(values) != 6 {
		e = errors.New("invalid PHC")
		return nil, nil, nil, e
	}

	var version int
	_, e = fmt.Sscanf(values[2], "v=%d", &version)
	if e != nil {
		return nil, nil, nil, e
	}
	if version != argon2.Version {
		e = errors.New("invalid version")
		return nil, nil, nil, e
	}

	p = &hashParams{}
	_, e = fmt.Sscanf(values[3], "m=%d,t=%d,p=%d", &p.memory, &p.iterations, &p.threads)
	if e != nil {
		e = errors.New("invalid parameters")
		return nil, nil, nil, e
	}

	s, e = base64.RawStdEncoding.Strict().DecodeString(values[4])
	if e != nil {
		return nil, nil, nil, e
	}
	p.saltLen = uint32(len(s))

	h, e = base64.RawStdEncoding.Strict().DecodeString(values[5])
	if e != nil {
		return nil, nil, nil, e
	}
	p.keyLen = uint32(len(h))

	return p, s, h, nil
}

func GenerateSalt(s uint32) ([]byte, error) {
	b := make([]byte, s)
	_, e := rand.Read(b)
	if e != nil {
		return nil, e
	}

	return b, nil
}
