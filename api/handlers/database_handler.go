package handlers

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// TODO switch to cfg baseURL
const BaseURL = "http://localhost:8000"

type DatabaseInterface interface {
	Exec(c context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(c context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(c context.Context, sql string, args ...interface{}) pgx.Row
	Begin(c context.Context) (pgx.Tx, error)
}

type DatabaseHandler struct {
	DB      DatabaseInterface
	Context *context.Context
}

func NewDatabaseHandler(database DatabaseInterface) *DatabaseHandler {
	return &DatabaseHandler{
		DB: database,
	}
}

type Loader struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}
