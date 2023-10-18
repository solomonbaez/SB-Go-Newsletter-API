package authentication

import (
	"context"
	"errors"
	"fmt"

	"strings"

	"github.com/rs/zerolog/log"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/handlers"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/models"
)

// TODO Sanitize credentials again?
func ValidateCredentials(c context.Context, dh *handlers.DatabaseHandler, credentials *models.Credentials) (*string, error) {
	var id string
	var password_hash string

	var user_e error
	query := "SELECT id, password_hash FROM users WHERE username=$1"
	e := dh.DB.QueryRow(c, query, credentials.Username).Scan(&id, &password_hash)
	if e != nil {
		log.Error().
			Err(e).
			Msg("Invalid username")

		// prevent timing attacks!
		password_hash = models.BaseHash
		user_e = e
	}

	if e := models.ValidatePHC(credentials.Password, password_hash); e != nil {
		if user_e != nil {
			e = user_e
		}
		return nil, e
	}

	return &id, nil
}

func ParseField(n string) (string, error) {
	// injection check
	for _, r := range n {
		c := string(r)
		if strings.Contains(models.InvalidRunes, c) {
			return "", fmt.Errorf("invalid character in name: %v", c)
		}
	}

	// empty field check
	nTrim := strings.Trim(n, " ")
	if nTrim == "" {
		return "", errors.New("name cannot be empty or whitespace")
	}

	return n, nil
}
