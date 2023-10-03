package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

var users gin.Accounts

func (rh *RouteHandler) GetUsers(c *gin.Context) (gin.Accounts, error) {
	var username string
	var password string

	query := "SELECT username, password FROM users"
	rows, e := rh.DB.Query(c, query)
	if e != nil {
		log.Error().
			Err(e).
			Msg("Failed to fetch admin users")

		return nil, e
	}

	// TODO fix lazy code
	_, _ = pgx.ForEachRow(rows, []any{&username, &password}, func() error {
		users[username] = password
		return nil
	})

	return users, nil
}
