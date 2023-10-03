package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

func (rh *RouteHandler) GetUsers(c *gin.Context) (gin.Accounts, error) {
	var users gin.Accounts
	var username string
	var password string

	rows, e := rh.DB.Query(c, "SELECT username, password FROM users")
	if e != nil {
		log.Error().
			Err(e).
			Msg("Failed to fetch admin users")

		return nil, e
	}

	_, e = pgx.ForEachRow(rows, []any{&username, &password}, func() error {
		users[username] = password
		return nil
	})
	if e != nil {
		log.Error().
			Err(e).
			Msg("Failed to parse admin users")

		return nil, e
	}

	return users, nil
}

func (rh *RouteHandler) ValidateCredentials(c *gin.Context, u string, p string) (*string, error) {
	var id string
	query := "SELECT id FROM users WHERE username=$1 AND password=$2"
	e := rh.DB.QueryRow(c, query, u, p).Scan(&id)
	if e != nil {
		log.Error().
			Err(e).
			Msg("Failed to fetch admin users")

		return nil, e
	}

	return &id, nil
}
