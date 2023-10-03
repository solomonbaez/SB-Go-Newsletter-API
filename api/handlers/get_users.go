package handlers

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

type Credentials struct {
	Username string
	Password string
}

func (rh *RouteHandler) GetUsers(c *gin.Context) (gin.Accounts, error) {
	var users gin.Accounts
	var username string
	var password string

	requestID := c.GetString("requestID")

	rows, e := rh.DB.Query(c, "SELECT username, password FROM users")
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

func (rh *RouteHandler) ValidateCredentials(c *gin.Context, username string, password string) (*string, error) {
	var id string

	requestID := c.GetString("requestID")

	query := "SELECT id FROM users WHERE username=$1 AND password=$2"
	e := rh.DB.QueryRow(c, query, username, password).Scan(&id)
	if e != nil {
		log.Error().
			Str("requestID", requestID).
			Err(e).
			Msg("Failed to validate user credentials")

		return nil, e
	}

	log.Info().
		Str("requestID", requestID).
		Str("userID", id).
		Msg("Successfully validated user credentials")

	return &id, nil
}

func BasicAuth(c *gin.Context) (*gin.Accounts, error) {
	var response string
	var e error

	requestID := c.GetString("requestID")
	h := c.GetHeader("Authorization")

	encodedSegment, valid := strings.CutPrefix(h, "Basic ")
	if !valid {
		e := errors.New("authorization method is not Basic")
		response = "Incorrect authorization method"
		HandleError(c, requestID, e, response, http.StatusBadRequest)
		return nil, e
	}

	decodedSegment, e := base64.RawStdEncoding.DecodeString(encodedSegment)
	if e != nil {
		response = "Failed to decode header"
		HandleError(c, requestID, e, response, http.StatusBadRequest)
		return nil, e
	}

	valid = utf8.Valid(decodedSegment)
	if !valid {
		e = errors.New("invalid header encoding")
		response = "Invalid header"
		HandleError(c, requestID, e, response, http.StatusBadRequest)
		return nil, e
	}

	// valid header should only contain two segments
	utf8Segment := string(decodedSegment)
	s := strings.Split(utf8Segment, ":")
	username := s[0]
	password := s[1]

	credentials := make(gin.Accounts)
	credentials[username] = password

	return &credentials, nil
}
