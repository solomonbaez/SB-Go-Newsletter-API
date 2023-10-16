package idempotency

import (
	"fmt"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"github.com/solomonbaez/SB-Go-Newsletter-API/api/handlers"
)

type NextAction struct {
	StartProcessing bool
	SavedResponse   *http.Response
}

func TryProcessing(c *gin.Context, dh *handlers.DatabaseHandler) (*NextAction, error) {
	session := sessions.Default(c)
	id := fmt.Sprintf("%s", session.Get("user"))
	key := fmt.Sprintf("%s", session.Get("key"))

	query := "INSERT INTO idempotency (id, idempotency_key, created) VALUES ($1, $2, now()) ON CONFLICT DO NOTHING"
	rows, e := dh.DB.Exec(c, query, id, key)
	if e != nil {
		return nil, e
	}

	if rows.RowsAffected() > 0 {
		return &NextAction{StartProcessing: true}, nil
	}

	savedResponse, e := GetSavedResponse(c, dh, id, key)
	if e != nil {
		return nil, e
	}

	return &NextAction{SavedResponse: savedResponse}, nil
}
