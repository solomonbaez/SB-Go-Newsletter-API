package handlers

import (
	"crypto/rand"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/models"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, "OK")
}

func StoreToken(c *gin.Context, tx pgx.Tx, id string, token string) error {
	query := "INSERT INTO subscription_tokens (subscription_token, subscriber_id) VALUES ($1, $2)"
	_, e := tx.Exec(c, query, token, id)
	if e != nil {
		return e
	}

	// commit changes
	tx.Commit(c)
	return nil
}

func GenerateCSPRNG(tokenLen int) (string, error) {
	b := make([]byte, tokenLen)

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
