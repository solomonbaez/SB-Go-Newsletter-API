package idempotency

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"

	"github.com/solomonbaez/SB-Go-Newsletter-API/api/handlers"
)

type HeaderPair struct {
	Name  string
	Value []byte
}

func GetSavedResponse(c *gin.Context, dh *handlers.DatabaseHandler, id, key string) (*http.Response, error) {
	var code int
	var headerPairRecord []HeaderPair
	var body []byte
	var e error

	tx, e := dh.DB.Begin(c)
	if e != nil {
		return nil, e
	}

	query := "SELECT response_status_code, response_body FROM idempotency WHERE user_id = $1 AND idempotency_key = $2"
	e = tx.QueryRow(c, query, id, key).Scan(&code, &body)
	if e != nil {
		return nil, e
	}

	query = "SELECT header_name, header_value FROM idempotency_headers WHERE idempotency_key = $1"
	e = tx.QueryRow(c, query, key).Scan(&headerPairRecord)
	if e != nil {
		return nil, e
	}

	// Construct the response
	response := &http.Response{
		Status:        http.StatusText(code),
		StatusCode:    code,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Header:        make(http.Header),
		ContentLength: -1, // Set the content length as needed
	}
	for _, hp := range headerPairRecord {
		response.Header.Set(hp.Name, string(hp.Value))
	}
	response.Body = io.NopCloser(bytes.NewReader(body))

	return response, nil
}

func SaveResponse(c *gin.Context, dh *handlers.DatabaseHandler, response *http.Response) error {
	var query string
	var e error
	var headerPairRecord []HeaderPair

	session := sessions.Default(c)
	id := session.Get("user")
	key := session.Get("key")

	status := uint16(response.StatusCode)

	for key, values := range response.Header {
		for _, value := range values {
			log.Info().
				Str("header", key).
				Msg("")

			headerPair := HeaderPair{Name: key, Value: []byte(value)}

			headerPairRecord = append(headerPairRecord, headerPair)
		}
	}

	tx, e := dh.DB.Begin(c)
	if e != nil {
		return e
	}
	defer tx.Rollback(c)

	// Read the response body into a byte slice
	bodyBytes, e := io.ReadAll(response.Body)
	if e != nil {
		return e
	}

	query = "INSERT INTO idempotency (id, idempotency_key, response_status_code, response_body, created) VALUES ($1, $2, $3, $4, now())"
	_, e = tx.Exec(c, query, id, key, status, bodyBytes)
	if e != nil {
		return e
	}

	query = "INSERT INTO idempotency_headers (idempotency_key, header_name, header_value) VALUES ($1, $2, $3)"
	for _, hp := range headerPairRecord {
		_, e = tx.Exec(c, query, key, hp.Name, hp.Value)
		if e != nil {
			return e
		}
	}
	if e := tx.Commit(c); e != nil {
		return e
	}

	return nil
}

func GetSavedResponses(c *gin.Context, dh *handlers.DatabaseHandler) {
	var savedResponses []*SavedResponse

	rows, e := dh.DB.Query(c, "SELECT * FROM idempotency")
	if e != nil {
		handlers.HandleError(c, "", e, "Query error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	savedResponses, e = pgx.CollectRows[*SavedResponse](rows, BuildResponse)
	if e != nil {
		handlers.HandleError(c, "", e, "Collect error", http.StatusInternalServerError)
	}

	c.JSON(http.StatusOK, gin.H{"responses": savedResponses})
}

type SavedResponse struct {
	Id      string    `json:"id"`
	Key     string    `json:"key"`
	Status  int       `json:"status"`
	Body    []byte    `json:"body"`
	Created time.Time `json:"created"`
}

func BuildResponse(row pgx.CollectableRow) (*SavedResponse, error) {
	var id string
	var key string
	var status int
	var body []byte
	var created time.Time

	e := row.Scan(&id, &key, &status, &body, &created)
	s := &SavedResponse{
		id,
		key,
		status,
		body,
		created,
	}

	return s, e
}
