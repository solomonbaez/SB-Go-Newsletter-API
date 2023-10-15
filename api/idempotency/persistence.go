package idempotency

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgtype"

	"github.com/solomonbaez/SB-Go-Newsletter-API/api/handlers"
)

type HeaderPair struct {
	name  string
	value uint8
}

func (hp *HeaderPair) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
	if src == nil {
		return errors.New("NULL values can't be decoded. Scan into a &*MyType to handle NULLs")
	}

	if e := (pgtype.CompositeFields{&hp.name, &hp.value}).DecodeBinary(ci, src); e != nil {
		return e
	}

	return nil
}

func GetSavedResponse(c *gin.Context, dh *handlers.DatabaseHandler, id, key string) (*http.Response, error) {
	var code int
	var protoHeader HeaderPair
	var body []byte

	query := "SELECT response_status_code, response_headers, response_body FROM idempotency WHERE user_id = $1 AND idempotency_key = $2"

	e := dh.DB.QueryRow(c, query, id, key).Scan(&code, &protoHeader, &body)
	if e != nil {
		return nil, e
	}

	// Construct the response
	proto := fmt.Sprintf("HTTP/1.%d", protoHeader.value)
	response := &http.Response{
		StatusCode: code,
		Proto:      proto,
	}

	response.Body = io.NopCloser(bytes.NewReader(body))

	return response, nil
}

func SaveResponse(c *gin.Context, dh *handlers.DatabaseHandler, response *http.Response) error {
	session := sessions.Default(c)
	id := session.Get("user")
	key := session.Get("key")

	status := uint16(response.StatusCode)
	headers := response.Header
	body := response.Body

	var e error
	var hp HeaderPair
	var headerPairRecord []HeaderPair
	for key, values := range headers {
		for _, value := range values {
			hp.name = key
			rawValue, e := strconv.Atoi(value)
			if e != nil {
				return e
			}
			hp.value = uint8(rawValue)

			headerPairRecord = append(headerPairRecord, hp)
		}
	}

	// Read the response body into a byte slice
	bodyBytes, e := io.ReadAll(body)
	if e != nil {
		return e
	}

	query := "INSERT id, idempotency_key, response_status_code, response_headers, response_body, created INTO idempotency VALUES ($1, $2, $3, $4, $5, now())"
	_, e = dh.DB.Exec(c, query, id, key, status, headerPairRecord, bodyBytes)
	if e != nil {
		return e
	}

	return e
}
