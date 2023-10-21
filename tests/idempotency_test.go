package api_test

import (
	"net/http"

	"testing"

	"github.com/pashagolub/pgxmock/v3"
	utils "github.com/solomonbaez/SB-Go-Newsletter-API/test_utils"
)

func Test_GetSavedResponse_Passes(t *testing.T) {
	app := utils.NewMockApp()
	defer app.Database.Close(app.Context)

	request, _ := http.NewRequest("GET", "/responses", nil)

	code := http.StatusSeeOther
	body := []byte("httpbody")

	query := "SELECT response_status_code, response_body FROM idempotency WHERE user_id = $1 AND idempotency_key = $2"
	app.Database.ExpectQuery(query).
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnRows(
			pgxmock.NewRows([]string{"respons_status_code", "response_body"}).
				AddRow(code, body),
		)

	name := "location"
	value := []byte("/admin/dashboard")
	query = "SELECT header_name, header_value FROM idempotency_headers WHERE idempotency_key = $1"
	app.Database.ExpectQuery(query).
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(
			pgxmock.NewRows([]string{"header_name", "header_value"}).
				AddRow(name, value),
		)

	app.NewMockRequest(request)
	app.Database.ExpectationsWereMet()
}

func Test_PostSaveResponse_Passes(t *testing.T) {
	app := utils.NewMockApp()
	defer app.Database.Close(app.Context)

	request, _ := http.NewRequest("POST", "/responses", nil)

	query := "UPDATE idempotency SET response_status_code = $3, response_body = $4 WHERE user_id = $1 AND idempotency_key = $2"
	app.Database.ExpectExec(query).
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	query = "UPDATE idempotency_headers SET header_name = $2, header_value = $3 WHERE idempotency_key = $1"
	app.Database.ExpectExec(query).
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	app.NewMockRequest(request)
	app.Database.ExpectationsWereMet()
}
