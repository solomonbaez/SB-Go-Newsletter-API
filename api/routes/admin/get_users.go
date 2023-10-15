package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/handlers"
)

func GetUsers(c *gin.Context, dh *handlers.DatabaseHandler) (gin.Accounts, error) {
	var users gin.Accounts
	var username string
	var password string

	requestID := c.GetString("requestID")

	rows, e := dh.DB.Query(c, "SELECT username, password FROM users")
	if e != nil {
		log.Error().
			Str("requestID", requestID).
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
			Str("requestID", requestID).
			Err(e).
			Msg("Failed to parse admin users")

		return nil, e
	}

	log.Info().
		Str("requestID", requestID).
		Msg("Successfully fetched admin users")

	return users, nil
}
