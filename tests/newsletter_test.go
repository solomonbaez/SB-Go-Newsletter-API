package api_test

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pashagolub/pgxmock/v3"

	"github.com/solomonbaez/SB-Go-Newsletter-API/api/models"
	adminRoutes "github.com/solomonbaez/SB-Go-Newsletter-API/api/routes/admin"
)

func TestPostNewsletter(t *testing.T) {
	testCases := &[]struct {
		name           string
		content        *models.Body
		expectedStatus int
		expectedHeader string
	}{
		{
			"(+) Test case 1 -> POST request to /admin/newsletter with valid content -> passes",
			&models.Body{
				Title: "test",
				Text:  "test",
				Html:  "<p>test</p>",
			},
			http.StatusSeeOther,
			"Newsletter",
		},
		{
			"(-) Test case 2 -> POST request to /admin/newsletter with invalid field -> fails",
			&models.Body{
				Title: "",
				Text:  "test",
				Html:  "<p>test</p>",
			},
			http.StatusBadRequest,
			"",
		},
	}

	// parallelize tests
	t.Parallel()
	var app App
	for _, tc := range *testCases {
		// initialize
		app = new_mock_app()
		admin = app.router.Group("/admin")
		admin.POST("/newsletter", func(c *gin.Context) { adminRoutes.PostNewsletter(c, app.dh, app.client) })
		defer app.database.Close(app.context)

		// Create a URL-encoded form data string
		data := url.Values{}
		data.Set("title", tc.content.Title)
		data.Set("text", tc.content.Text)
		data.Set("html", tc.content.Title)
		formData := data.Encode()

		// Create a POST request with the form data
		request, _ := http.NewRequest("POST", "/admin/newsletter", strings.NewReader(formData))
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		app.database.ExpectBegin()

		query := "INSERT INTO idempotency"
		app.database.ExpectExec(query).
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		query = "INSERT INTO idempotency_headers"
		app.database.ExpectExec(query).
			WithArgs(pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		query = "INSERT INTO newsletter_issues"
		app.database.ExpectExec(query).
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		query = "INSERT INTO issue_delivery_queue"
		app.database.ExpectExec(query).
			WithArgs(pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		app.database.ExpectCommit()
		app.database.ExpectBegin()

		query = "UPDATE idempotency SET"
		app.database.ExpectExec(query).
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		query = "UPDATE idempotency_headers SET"
		app.database.ExpectExec(query).
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		app.database.ExpectCommit()

		app.new_mock_request(request)
		defer app.database.ExpectationsWereMet()

		// tests
		if responseStatus := app.recorder.Code; responseStatus != tc.expectedStatus {
			t.Errorf("Expected status code %v, but got %v", tc.expectedStatus, responseStatus)
		}
		responseHeader := app.recorder.Header().Get("X-Redirect")
		if responseHeader != tc.expectedHeader {
			t.Errorf("Expected header %s, but got %s", tc.expectedHeader, responseHeader)
		}
	}
}
