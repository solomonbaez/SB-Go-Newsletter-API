package api_test

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pashagolub/pgxmock/v3"

	"github.com/solomonbaez/hyacinth/api/models"
	adminRoutes "github.com/solomonbaez/hyacinth/api/routes/admin"
	utils "github.com/solomonbaez/hyacinth/test_utils"
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
	var app utils.App
	for _, tc := range *testCases {
		// initialize
		app = utils.NewMockApp()
		admin := app.Router.Group("/admin")
		admin.POST("/newsletter", func(c *gin.Context) { adminRoutes.PostNewsletter(c, app.DH, app.Client) })
		defer app.Database.Close(app.Context)

		// Create a URL-encoded form data string
		data := url.Values{}
		data.Set("title", tc.content.Title)
		data.Set("text", tc.content.Text)
		data.Set("html", tc.content.Title)
		formData := data.Encode()

		// Create a POST request with the form data
		request, _ := http.NewRequest("POST", "/admin/newsletter", strings.NewReader(formData))
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		app.Database.ExpectBegin()

		query := "INSERT INTO idempotency"
		app.Database.ExpectExec(query).
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		query = "INSERT INTO idempotency_headers"
		app.Database.ExpectExec(query).
			WithArgs(pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		query = "INSERT INTO newsletter_issues"
		app.Database.ExpectExec(query).
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		query = "INSERT INTO issue_delivery_queue"
		app.Database.ExpectExec(query).
			WithArgs(pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		app.Database.ExpectCommit()
		app.Database.ExpectBegin()

		query = "UPDATE idempotency SET"
		app.Database.ExpectExec(query).
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		query = "UPDATE idempotency_headers SET"
		app.Database.ExpectExec(query).
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		app.Database.ExpectCommit()

		app.NewMockRequest(request)
		defer app.Database.ExpectationsWereMet()

		// tests
		if responseStatus := app.Recorder.Code; responseStatus != tc.expectedStatus {
			t.Errorf("Expected status code %v, but got %v", tc.expectedStatus, responseStatus)
		}
		responseHeader := app.Recorder.Header().Get("X-Redirect")
		if responseHeader != tc.expectedHeader {
			t.Errorf("Expected header %s, but got %s", tc.expectedHeader, responseHeader)
		}
	}
}
