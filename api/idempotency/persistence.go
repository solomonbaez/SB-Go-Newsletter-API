package idempotency

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/solomonbaez/SB-Go-Newsletter-API/api/handlers"
)

type HeaderPair struct {
	Name  string
	Value []byte
}

func GetSavedResponse(c context.Context, dh *handlers.DatabaseHandler, id, key string) (*http.Response, error) {
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

func SaveResponse(c context.Context, dh *handlers.DatabaseHandler, id, key string, response *http.Response) error {
	var query string
	var e error
	var headerPairRecord []HeaderPair

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

	query = "UPDATE idempotency SET response_status_code = $3, response_body = $4 WHERE id = $1 AND idempotency_key = $2"
	_, e = tx.Exec(c, query, id, key, status, bodyBytes)
	if e != nil {
		return e
	}

	query = "UPDATE idempotency_headers SET header_name = $2, header_value = $3 WHERE idempotency_key = $1"
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
