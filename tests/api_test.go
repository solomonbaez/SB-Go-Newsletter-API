package api_test

import (
	"errors"
	"net/http"

	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v3"

	"github.com/solomonbaez/hyacinth/api/handlers"
	"github.com/solomonbaez/hyacinth/api/models"
	"github.com/solomonbaez/hyacinth/api/routes"
	adminRoutes "github.com/solomonbaez/hyacinth/api/routes/admin"
	utils "github.com/solomonbaez/hyacinth/test_utils"
)

func TestHealthCheck(t *testing.T) {
	test := &struct {
		name           string
		expectedStatus int
	}{
		"(+) Test case -> GET request to /health -> passes",
		http.StatusOK,
	}

	t.Parallel()
	// initialize
	app := utils.NewMockApp()
	app.Router.GET("/health", handlers.HealthCheck)
	defer app.Database.Close(app.Context)

	request, _ := http.NewRequest("GET", "/health", nil)

	// assertions
	app.NewMockRequest(request)
	if responseStatus := app.Recorder.Code; responseStatus != test.expectedStatus {
		t.Errorf("Expected status code %v, but got %v", http.StatusOK, responseStatus)
	}

	expected_body := `"OK"`
	response_body := app.Recorder.Body.String()
	if response_body != expected_body {
		t.Errorf("Expected body %v, but got %v", expected_body, response_body)
	}
}

func TestGetSubscribers(t *testing.T) {
	seedSubscriber := &struct {
		id      string
		email   models.SubscriberEmail
		name    models.SubscriberName
		created time.Time
		status  string
	}{
		uuid.NewString(),
		models.SubscriberEmail("user@example.com"),
		models.SubscriberName("user"),
		time.Now(),
		"pending",
	}

	testCases := &[]struct {
		name           string
		subscribers    bool
		expectedStatus int
		expectedBody   string
	}{
		{
			"(+) Test case 1 -> GET request to /subscribers with no subscribers -> passes",
			false,
			http.StatusOK,
			`{"requestID":"","subscribers":[]}`,
		},
		{
			"(+) Test case 2 -> GET request to /subscribers with subscribers -> passes",
			true,
			http.StatusOK,
			fmt.Sprintf(
				`{"requestID":"","subscribers":[{"id":"%s","email":"%s","name":"%s","status":"%s"}]}`,
				seedSubscriber.id,
				seedSubscriber.email,
				seedSubscriber.name,
				seedSubscriber.status,
			),
		},
	}

	t.Parallel()
	for _, tc := range *testCases {
		// initialize
		app := utils.NewMockApp()
		admin := app.Router.Group("/admin")
		admin.GET("/subscribers", func(c *gin.Context) { adminRoutes.GetSubscribers(c, app.DH) })
		defer app.Database.Close(app.Context)

		request, _ := http.NewRequest("GET", "/admin/subscribers", nil)

		if tc.subscribers {
			app.Database.ExpectQuery(`SELECT \* FROM subscriptions`).
				WillReturnRows(
					pgxmock.NewRows([]string{"id", "email", "name", "created", "status"}).
						AddRow(
							seedSubscriber.id,
							seedSubscriber.email,
							seedSubscriber.name,
							seedSubscriber.created,
							seedSubscriber.status,
						),
				)
		} else {
			app.Database.ExpectQuery(`SELECT \* FROM subscriptions`).
				WillReturnRows(
					pgxmock.NewRows([]string{"id", "email", "name", "created", "status"}),
				)
		}

		app.NewMockRequest(request)
		defer app.Database.ExpectationsWereMet()

		// tests
		if responseStatus := app.Recorder.Code; responseStatus != tc.expectedStatus {
			t.Errorf("Expected status code %v, but got %v", tc.expectedStatus, responseStatus)
		}

		responseBody := app.Recorder.Body.String()
		if responseBody != tc.expectedBody {
			t.Errorf("Expected body %v, but got %v", tc.expectedBody, responseBody)
		}
	}
}

func TestGetConfirmedSubscribers(t *testing.T) {
	seedSubscriber := &struct {
		id      string
		email   models.SubscriberEmail
		name    models.SubscriberName
		created time.Time
		status  string
	}{
		uuid.NewString(),
		models.SubscriberEmail("user@example.com"),
		models.SubscriberName("user"),
		time.Now(),
		"confirmed",
	}

	test := &struct {
		name          string
		expectedArray []*models.Subscriber
	}{
		"(+) Test case -> -> passes",
		[]*models.Subscriber{
			{
				ID:     seedSubscriber.id,
				Email:  seedSubscriber.email,
				Name:   seedSubscriber.name,
				Status: seedSubscriber.status,
			},
		},
	}

	t.Parallel()
	// initialize
	app := utils.NewMockApp()
	defer app.Database.Close(app.Context)

	app.Database.ExpectQuery(`SELECT id, email, name, created, status FROM subscriptions WHERE`).
		WithArgs(seedSubscriber.status).
		WillReturnRows(
			pgxmock.NewRows([]string{"id", "email", "name", "created", "status"}).
				AddRow(
					seedSubscriber.id,
					seedSubscriber.email,
					seedSubscriber.name,
					seedSubscriber.created,
					seedSubscriber.status,
				),
		)

	responseArray := adminRoutes.GetConfirmedSubscribers(app.Context, app.DH)
	defer app.Database.ExpectationsWereMet()

	if *responseArray[0] != *test.expectedArray[0] {
		t.Errorf("Expected array: %v, got: %v", *test.expectedArray[0], *responseArray[0])
	}
}

func TestGetSubscribersByID(t *testing.T) {
	seedSubscriber := &struct {
		id      string
		email   models.SubscriberEmail
		name    models.SubscriberName
		created time.Time
		status  string
	}{
		uuid.NewString(),
		models.SubscriberEmail("user@example.com"),
		models.SubscriberName("user"),
		time.Now(),
		"pending",
	}

	testCases := &[]struct {
		name           string
		validID        bool
		expectedStatus int
		expectedBody   string
	}{
		{
			"(+) Test case -> GET to /admin/subscribers/:id with valid ID -> passes",
			true,
			http.StatusFound,
			fmt.Sprintf(
				`{"requestID":"","subscriber":{"id":"%s","email":"%s","name":"%s","status":"%s"}}`,
				seedSubscriber.id,
				seedSubscriber.email.String(),
				seedSubscriber.name.String(),
				seedSubscriber.status,
			),
		},
		{
			"(-) Test case -> GET to /admin/subscribers/:id with invalid ID -> fails",
			false,
			http.StatusNotFound,
			`{"error":"Database query error: Invalid ID","requestID":""}`,
		},
	}

	t.Parallel()
	for _, tc := range *testCases {
		// initialization
		app := utils.NewMockApp()
		admin := app.Router.Group("/admin")
		admin.GET("/subscribers/:id", func(c *gin.Context) { adminRoutes.GetSubscriberByID(c, app.DH) })
		defer app.Database.Close(app.Context)

		request, _ := http.NewRequest("GET", fmt.Sprintf("/admin/subscribers/%v", seedSubscriber.id), nil)

		query := app.Database.ExpectQuery(`SELECT id, email, name, status FROM subscriptions WHERE`).
			WithArgs(pgxmock.AnyArg())
		if tc.validID {
			query.WillReturnRows(
				pgxmock.NewRows([]string{"id", "email", "name", "status"}).
					AddRow(
						seedSubscriber.id,
						seedSubscriber.email,
						seedSubscriber.name,
						seedSubscriber.status,
					),
			)
		} else {
			query.WillReturnError(errors.New("Invalid ID"))
		}

		// tests
		app.NewMockRequest(request)
		defer app.Database.ExpectationsWereMet()

		if responseStatus := app.Recorder.Code; responseStatus != tc.expectedStatus {
			t.Errorf("Expected status code %v, but got %v", tc.expectedStatus, responseStatus)
		}

		responseBody := app.Recorder.Body.String()
		if responseBody != tc.expectedBody {
			t.Errorf("Expected body %v, but got %v", tc.expectedBody, responseBody)
		}
	}
}

func TestPostSubscribe(t *testing.T) {
	seedSubscriber := &struct {
		id      string
		email   models.SubscriberEmail
		name    models.SubscriberName
		created time.Time
		status  string
	}{
		uuid.NewString(),
		models.SubscriberEmail("user@example.com"),
		models.SubscriberName("user"),
		time.Now(),
		"pending",
	}

	testCases := &[]struct {
		name           string
		data           []string
		expectedStatus int
	}{
		{
			"(+) Test case -> POST to /subscribe with valid fields -> passes",

			[]string{
				fmt.Sprintf(
					`{"email": "%s", "name": "%s"}`,
					seedSubscriber.email.String(),
					seedSubscriber.name.String(),
				),
			},

			http.StatusCreated,
		},
		{
			"(-) Test case -> POST to /subscribe with invalid email -> fails",
			[]string{
				`{email: "", "name": "user"}`,
				`{email: " ", "name": "user"}`,
				`{"email": "user", "name": "user"}`,
				`{"email": "user@", "name": "user"}`,
				`{"email": "@example.com", "name": "user"}`,
				`{"email": "example.com", "name": "user"}`,
			},

			http.StatusBadRequest,
		},
		{
			"(-) Test case -> POST to /subscribe with invalid name -> fails",
			[]string{
				`{"email": "user@example.com", "name": ""}`,
				`{"email": "user@example.com", "name": " "}`,
				`{"email": "user@example.com", "name": "user{"}`,
				`{"email": "user@example.com, "name": "user}"}`,
				`{"email": "user@example.com", "name": "user/"}`,
				`{"email": "user@example.com", "name": "user\\"}`,
				`{"email": "user@example.com", "name": "user<"}`,
				`{"email": "user@example.com", "name": "user>"}`,
				`{"email": "user@example.com", "name": "user("}`,
				`{"email": "user@example.com", "name": "user)"}`,
			},

			http.StatusBadRequest,
		},
	}

	t.Parallel()

	for _, tc := range *testCases {

		for _, d := range tc.data {
			// initialization
			app := utils.NewMockApp()
			app.Router.POST("/subscribe", func(c *gin.Context) { routes.Subscribe(c, app.DH) })
			request, _ := http.NewRequest("POST", "/subscribe", strings.NewReader(d))

			app.Database.ExpectBegin()
			app.Database.ExpectExec("INSERT INTO subscriptions").
				WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
				WillReturnResult(pgxmock.NewResult("INSERT", 1))
			app.Database.ExpectExec("INSERT INTO subscription_tokens").
				WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
				WillReturnResult(pgxmock.NewResult("INSERT", 1))
			app.Database.ExpectExec("INSERT INTO issue_delivery_queue").
				WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
				WillReturnResult(pgxmock.NewResult("INSERT", 1))
			app.Database.ExpectCommit()

			app.NewMockRequest(request)

			// tests
			if responseStatus := app.Recorder.Code; responseStatus != tc.expectedStatus {
				t.Errorf("Expected status code %v, but got %v", tc.expectedStatus, responseStatus)
			}

			app.Database.ExpectationsWereMet()
			app.Database.Close(app.Context)
		}
	}
}

func TestConfirmSubscriber(t *testing.T) {
	seedSubscriber := &struct {
		token string
		id    string
	}{
		uuid.NewString(),
		uuid.NewString(),
	}

	testCases := &[]struct {
		name           string
		token          string
		expectedStatus int
		expectedBody   string
	}{
		{
			"(+) Test case -> POST to /confirm/:token with valid id -> passes",
			seedSubscriber.token,
			http.StatusAccepted,
			`{"requestID":"","subscriber":"Subscription confirmed"}`,
		},
		{
			"(-) Test case -> POST to /confirm/:token with invalid id -> fails",
			uuid.NewString(),
			http.StatusInternalServerError,
			`{"error":"Failed to fetch subscriber ID: invalid token","requestID":""}`,
		},
	}

	for _, tc := range *testCases {
		// initialize
		app := utils.NewMockApp()
		app.Router.GET("/confirm/:token", func(c *gin.Context) { routes.ConfirmSubscriber(c, app.DH) })
		defer app.Database.Close(app.Context)

		request, _ := http.NewRequest("GET", fmt.Sprintf("/confirm/%s", tc.token), nil)

		query := app.Database.ExpectQuery(`SELECT subscriber_id FROM subscription_tokens WHERE`).
			WithArgs(pgxmock.AnyArg())
		if tc.token == seedSubscriber.token {
			query.WillReturnRows(
				pgxmock.NewRows([]string{"subscriber_id"}).
					AddRow(seedSubscriber.id),
			)
		} else {
			query.WillReturnError(errors.New("invalid token"))
		}

		app.Database.ExpectExec(`UPDATE subscriptions SET status = 'confirmed' WHERE`).
			WithArgs(pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		app.NewMockRequest(request)
		defer app.Database.ExpectationsWereMet()

		// tests
		if responseStatus := app.Recorder.Code; responseStatus != tc.expectedStatus {
			t.Errorf("Expected status code %v, but got %v", tc.expectedStatus, responseStatus)
		}

		responseBody := app.Recorder.Body.String()
		if responseBody != tc.expectedBody {
			t.Errorf("Expected body %v, but got %v", tc.expectedBody, responseBody)
		}
	}
}
