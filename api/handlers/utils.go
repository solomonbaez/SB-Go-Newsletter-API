package handlers

import (
	"context"
	"crypto/rand"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rs/zerolog/log"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/models"
)

// TODO switch to cfg baseURL
const baseURL = "http://localhost:8000"
const tokenLength = 25
const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type Database interface {
	Exec(c context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(c context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(c context.Context, sql string, args ...interface{}) pgx.Row
	Begin(c context.Context) (pgx.Tx, error)
}

type RouteHandler struct {
	DB Database
}

func NewRouteHandler(db Database) *RouteHandler {
	return &RouteHandler{
		DB: db,
	}
}

type Loader struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, "OK")
}

func storeToken(c *gin.Context, tx pgx.Tx, id string, token string) error {
	query := "INSERT INTO subscription_tokens (subscription_token, subscriber_id) VALUES ($1, $2)"
	_, e := tx.Exec(c, query, token, id)
	if e != nil {
		return e
	}

	// commit changes
	tx.Commit(c)
	return nil
}

func GenerateCSPRNG() (string, error) {
	b := make([]byte, tokenLength)

	maxIndex := big.NewInt(int64(len(charset)))

	for i := range b {
		r, e := rand.Int(rand.Reader, maxIndex)
		if e != nil {
			return "", e
		}

		b[i] = charset[r.Int64()]
	}

	return string(b), nil
}

func BuildSubscriber(row pgx.CollectableRow) (*models.Subscriber, error) {
	var id string
	var email models.SubscriberEmail
	var name models.SubscriberName
	var created time.Time
	var status string

	e := row.Scan(&id, &email, &name, &created, &status)
	s := &models.Subscriber{
		ID:     id,
		Email:  email,
		Name:   name,
		Status: status,
	}

	return s, e
}

func HandleError(c *gin.Context, id string, e error, response string, status int) {
	log.Error().
		Str("requestID", id).
		Err(e).
		Msg(response)

	var message strings.Builder
	message.WriteString(response)
	message.WriteString(": ")
	message.WriteString(e.Error())

	c.JSON(status, gin.H{"requestID": id, "error": message.String()})
}
