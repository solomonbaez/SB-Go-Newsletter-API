package authentication

import (
	"context"
	"errors"
	"fmt"

	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/handlers"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/models"
)

// TODO Sanitize credentials again?
func ValidateCredentials(c context.Context, dh *handlers.DatabaseHandler, credentials *models.Credentials) (id *string, err error) {
	var userID string
	var passwordHash string

	var user_e error
	query := "SELECT id, password_hash FROM users WHERE username=$1"
	e := dh.DB.QueryRow(c, query, credentials.Username).Scan(&userID, &passwordHash)
	if e != nil {
		if errors.Is(e, pgx.ErrNoRows) {
			e = errors.New("user not found")
		} else {
			e = fmt.Errorf("database error: %w", e)
		}

		log.Error().
			Err(e).
			Msg("Invalid username")

		// prevent timing attacks!
		passwordHash = models.BaseHash
		user_e = e
	}

	if e = models.ValidatePHC(credentials.Password, passwordHash); e != nil {
		if user_e != nil {
			e = user_e
		} else {
			e = fmt.Errorf("invalid credentials: %w", e)
		}
		return nil, e
	}

	id = &userID
	return id, nil
}

func ParseField(field string) (parsed *string, err error) {
	// injection check
	for _, r := range field {
		c := string(r)
		if strings.Contains(models.InvalidRunes, c) {
			return nil, fmt.Errorf("invalid character in field: %s", c)
		}
	}

	// empty field check
	trimmedField := strings.Trim(field, " ")
	if trimmedField == "" {
		return nil, errors.New("field cannot be empty or whitespace")
	}

	parsed = &field
	return parsed, nil
}
