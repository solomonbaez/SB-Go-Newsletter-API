package idempotency

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"

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
