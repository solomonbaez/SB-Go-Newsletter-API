package api_test

import (
	"net/http"

	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v3"

	"github.com/solomonbaez/SB-Go-Newsletter-API/api/handlers"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/models"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/routes"
	adminRoutes "github.com/solomonbaez/SB-Go-Newsletter-API/api/routes/admin"
)

func Test_HealthCheck_Returns_OK(t *testing.T) {
	// initialize
	app := new_mock_app()
	defer app.database.Close(app.context)

	app.router.GET("/health", handlers.HealthCheck)

	// server initialization
	request, e := http.NewRequest("GET", "/health", nil)
	if e != nil {
		t.Fatal(e)
	}

	// tests
	app.new_mock_request(request)
	if status := app.recorder.Code; status != http.StatusOK {
		t.Errorf("Expected status code %v, but got %v", http.StatusOK, status)
	}

	expected_body := `"OK"`
	response_body := app.recorder.Body.String()
	if response_body != expected_body {
		t.Errorf("Expected body %v, but got %v", expected_body, response_body)
	}
}

func Test_GetSubscribers_NoSubscribers_Passes(t *testing.T) {
	// initialize
	app := new_mock_app()
	defer app.database.Close(app.context)

	admin = app.router.Group("/admin")
	admin.GET("/subscribers", func(c *gin.Context) { adminRoutes.GetSubscribers(c, app.dh) })

	request, e := http.NewRequest("GET", "/admin/subscribers", nil)
	if e != nil {
		t.Fatal(e)
	}

	app.database.ExpectQuery(`SELECT \* FROM subscriptions`).WillReturnRows(
		pgxmock.NewRows([]string{"id", "email", "name", "created", "status"}),
	)

	app.new_mock_request(request)
	defer app.database.ExpectationsWereMet()

	// tests
	if status := app.recorder.Code; status != http.StatusOK {
		t.Errorf("Expected status code %v, but got %v", http.StatusOK, status)
	}

	expected_body := `{"requestID":"","subscribers":"No subscribers"}`
	response_body := app.recorder.Body.String()
	if response_body != expected_body {
		t.Errorf("Expected body %v, but got %v", expected_body, response_body)
	}
}

func Test_GetSubscribers_WithSubscribers_Passes(t *testing.T) {
	// initialization
	app := new_mock_app()
	defer app.database.Close(app.context)

	admin = app.router.Group("/admin")
	admin.GET("/subscribers", func(c *gin.Context) { adminRoutes.GetSubscribers(c, app.dh) })

	request, e := http.NewRequest("GET", "/admin/subscribers", nil)
	if e != nil {
		t.Fatal(e)
	}

	mock_id := uuid.NewString()
	app.database.ExpectQuery(`SELECT \* FROM subscriptions`).WillReturnRows(
		pgxmock.NewRows([]string{"id", "email", "name", "created", "status"}).
			AddRow(mock_id, models.SubscriberEmail("test@test.com"), models.SubscriberName("TestUser"), time.Now(), "pending"),
	)

	app.new_mock_request(request)
	defer app.database.ExpectationsWereMet()

	// tests
	if status := app.recorder.Code; status != http.StatusOK {
		t.Errorf("Expected status code %v, but got %v", http.StatusOK, status)
	}

	expected_body := fmt.Sprintf(`{"requestID":"","subscribers":[{"id":"%v","email":"test@test.com","name":"TestUser","status":"pending"}]}`, mock_id)
	response_body := app.recorder.Body.String()
	if response_body != expected_body {
		t.Errorf("Expected body %v, but got %v", expected_body, response_body)
	}
}

func Test_GetConfirmedSubscribers_NoSubscribers_Passes(t *testing.T) {
	// initialize
	app := new_mock_app()
	defer app.database.Close(app.context)

	admin = app.router.Group("/admin")
	admin.GET("/confirmed", func(c *gin.Context) { _ = adminRoutes.GetConfirmedSubscribers(c, app.dh) })

	request, e := http.NewRequest("GET", "/admin/confirmed", nil)
	if e != nil {
		t.Fatal(e)
	}

	app.database.ExpectQuery(`SELECT id, email, name, created, status FROM subscriptions WHERE`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(
			pgxmock.NewRows([]string{"id", "email", "name", "created", "status"}),
		)

	app.new_mock_request(request)
	defer app.database.ExpectationsWereMet()

	// tests
	if status := app.recorder.Code; status != http.StatusOK {
		t.Errorf("Expected status code %v, but got %v", http.StatusOK, status)
	}
}

func Test_GetConfirmedSubscribers_WithSubscribers_Passes(t *testing.T) {
	// initialize
	app := new_mock_app()
	defer app.database.Close(app.context)

	admin = app.router.Group("/admin")
	admin.GET("/confirmed", func(c *gin.Context) { _ = adminRoutes.GetConfirmedSubscribers(c, app.dh) })

	request, e := http.NewRequest("GET", "/admin/confirmed", nil)
	if e != nil {
		t.Fatal(e)
	}

	mock_id := uuid.NewString()
	app.database.ExpectQuery(`SELECT id, email, name, created, status FROM subscriptions WHERE`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(
			pgxmock.NewRows([]string{"id", "email", "name", "created", "status"}).
				AddRow(mock_id, models.SubscriberEmail("test@test.com"), models.SubscriberName("TestUser"), time.Now(), "confirmed"),
		)

	app.new_mock_request(request)
	defer app.database.ExpectationsWereMet()

	// tests
	if status := app.recorder.Code; status != http.StatusOK {
		t.Errorf("Expected status code %v, but got %v", http.StatusOK, status)
	}
}

func Test_GetSubscribersByID_ValidID_Passes(t *testing.T) {
	// initialization
	app := new_mock_app()
	defer app.database.Close(app.context)

	admin = app.router.Group("/admin")
	admin.GET("/subscribers/:id", func(c *gin.Context) { adminRoutes.GetSubscriberByID(c, app.dh) })

	mock_id := uuid.NewString()
	request, e := http.NewRequest("GET", fmt.Sprintf("/admin/subscribers/%v", mock_id), nil)
	if e != nil {
		t.Fatal(e)
	}

	app.database.ExpectQuery(`SELECT id, email, name, status FROM subscriptions WHERE`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(
			pgxmock.NewRows([]string{"id", "email", "name", "status"}).
				AddRow(mock_id, models.SubscriberEmail("test@test.com"), models.SubscriberName("TestUser"), "pending"),
		)

	// tests
	app.new_mock_request(request)
	defer app.database.ExpectationsWereMet()

	if status := app.recorder.Code; status != http.StatusFound {
		t.Errorf("Expected status code %v, but got %v", http.StatusFound, status)
	}

	expected_body := fmt.Sprintf(`{"requestID":"","subscriber":{"id":"%v","email":"test@test.com","name":"TestUser","status":"pending"}}`, mock_id)
	response_body := app.recorder.Body.String()

	if response_body != expected_body {
		t.Errorf("Expected body %v, but got %v", expected_body, response_body)
	}
}

func Test_GetSubscribersByID_InvalidID_Fails(t *testing.T) {
	// initialization
	app := new_mock_app()
	defer app.database.Close(app.context)

	admin = app.router.Group("/admin")
	admin.GET("/subscribers/:id", func(c *gin.Context) { adminRoutes.GetSubscriberByID(c, app.dh) })

	// Non-UUID ID
	mock_id := "1"
	request, e := http.NewRequest("GET", fmt.Sprintf("/admin/subscribers/%v", mock_id), nil)
	if e != nil {
		t.Fatal(e)
	}

	app.database.ExpectQuery(`SELECT id, email, name FROM subscriptions WHERE`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(
			pgxmock.NewRows([]string{"id", "email", "name"}).
				AddRow(mock_id, models.SubscriberEmail("test@test.com"), models.SubscriberName("TestUser")),
		)

	// tests
	app.new_mock_request(request)
	defer app.database.ExpectationsWereMet()

	if status := app.recorder.Code; status != http.StatusBadRequest {
		t.Errorf("Expected status code %v, but got %v", http.StatusBadRequest, status)
	}

	expected_body := `{"error":"Invalid ID format: invalid UUID length: 1","requestID":""}`
	response_body := app.recorder.Body.String()

	if response_body != expected_body {
		t.Errorf("Expected body %v, but got %v", expected_body, response_body)
	}
}

func Test_Subscribe_Passes(t *testing.T) {
	// initialization
	app := new_mock_app()
	defer app.database.Close(app.context)

	app.router.POST("/subscribe", func(c *gin.Context) { routes.Subscribe(c, app.dh, app.client) })

	data := `{"email": "test@test.com", "name": "TestUser"}`
	request, e := http.NewRequest("POST", "/subscribe", strings.NewReader(data))
	if e != nil {
		t.Fatal(e)
	}

	app.database.ExpectBegin()
	app.database.ExpectExec("INSERT INTO subscriptions").
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	app.database.ExpectExec("INSERT INTO subscription_tokens").
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	app.database.ExpectCommit()

	app.new_mock_request(request)
	defer app.database.ExpectationsWereMet()

	// tests
	if status := app.recorder.Code; status != http.StatusCreated {
		t.Errorf("Expected status code %v, but got %v", http.StatusCreated, status)
	}

	expected_body := `{"requestID":"","subscriber":{"id":"","email":"test@test.com","name":"TestUser","status":"pending"}}`
	response_body := app.recorder.Body.String()
	if response_body != expected_body {
		t.Errorf("Expected body %v, but got %v", expected_body, response_body)
	}
}

func Test_Subscribe_InvalidEmail_Fails(t *testing.T) {
	// // initialization
	var app App
	var request *http.Request
	var e error

	var test_cases []string
	test_cases = append(test_cases,
		`{email: "", "name": "TestUser"}`,
		`{email: " ", "name": "TestUser"}`,
		`{"email": "test", "name": "TestUser"}`,
		`{"email": "test@", "name": "TestUser"}`,
		`{"email": "@test.com", "name": "TestUser"}`,
		`{"email": "test.com", "name": "TestUser"}`,
	)
	for _, tc := range test_cases {
		// resource intensive but necessary duplication
		app = new_mock_app()
		app.router.POST("/subscribe", func(c *gin.Context) { routes.Subscribe(c, app.dh, app.client) })

		request, e = http.NewRequest("POST", "/subscribe", strings.NewReader(tc))
		if e != nil {
			t.Fatal(e)
		}

		app.database.ExpectBegin()
		app.database.ExpectExec("INSERT INTO subscriptions").
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg())
		app.database.ExpectRollback()

		app.new_mock_request(request)

		// tests
		if status := app.recorder.Code; status != http.StatusBadRequest {
			t.Errorf("Expected status code %v, but got %v", http.StatusBadRequest, status)
		}

		app.database.ExpectationsWereMet()
		app.database.Close(app.context)
	}
}

func TestSubscribeInvalidNameFails(t *testing.T) {
	// // initialization
	var request *http.Request
	var app App
	var e error

	var test_cases []string
	test_cases = append(test_cases,
		`{"email": "test@email.com", "name": ""}`,
		`{"email": "test@email.com", "name": " "}`,
		`{"email": "test@email.com", "name": "test{"}`,
		`{"email": "test@email.com", "name": "test}"}`,
		`{"email": "test@email.com", "name": "test/"}`,
		`{"email": "test@email.com", "name": "test\\"}`,
		`{"email": "test@email.com", "name": "test<"}`,
		`{"email": "test@email.com", "name": "test>"}`,
		`{"email": "test@email.com", "name": "test("}`,
		`{"email": "test@email.com", "name": "test)"}`,
	)
	for _, tc := range test_cases {
		// resource intensive but necessary duplication
		app = new_mock_app()
		app.router.POST("/subscribe", func(c *gin.Context) { routes.Subscribe(c, app.dh, app.client) })

		request, e = http.NewRequest("POST", "/subscribe", strings.NewReader(tc))
		if e != nil {
			t.Fatal(e)
		}

		app.database.ExpectBegin()
		app.database.ExpectExec("INSERT INTO subscriptions").
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg())
		app.database.ExpectRollback()

		app.new_mock_request(request)

		// tests
		if status := app.recorder.Code; status != http.StatusBadRequest {
			t.Errorf("Expected status code %v, but got %v", http.StatusBadRequest, status)
		}

		app.database.ExpectationsWereMet()
		app.database.Close(app.context)
	}
}

func Test_Subscribe_MaxLengthParameters_Fails(t *testing.T) {
	// // initialization
	var request *http.Request
	var app App
	var e error

	long_email := "a" + strings.Repeat("a", 100) + "@test.com"
	long_name := "a" + strings.Repeat("a", 100)

	var test_cases []string
	test_cases = append(test_cases,
		fmt.Sprintf(`{"email": "%v", "name": "TestUser"}`, long_email),
		fmt.Sprintf(`{"email": "test@test.com", "name": "%v"}`, long_name),
		fmt.Sprintf(`{"email": "%v", "name": "%v"}`, long_email, long_name),
	)
	for _, tc := range test_cases {
		// resource intensive but necessary duplication
		app = new_mock_app()
		app.router.POST("/subscribe", func(c *gin.Context) { routes.Subscribe(c, app.dh, app.client) })

		request, e = http.NewRequest("POST", "/subscribe", strings.NewReader(tc))
		if e != nil {
			t.Fatal(e)
		}

		app.database.ExpectBegin()
		app.database.ExpectExec("INSERT INTO subscriptions").
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg())
		app.database.ExpectRollback()

		app.new_mock_request(request)

		// tests
		if status := app.recorder.Code; status != http.StatusBadRequest {
			t.Errorf("Expected status code %v, but got %v", http.StatusBadRequest, status)
		}

		app.database.ExpectationsWereMet()
		app.database.Close(app.context)
	}
}

func Test_ConfirmSubscriber_Passes(t *testing.T) {
	// initialize
	app := new_mock_app()
	defer app.database.Close(app.context)

	app.router.GET("/confirm/:token", func(c *gin.Context) { routes.ConfirmSubscriber(c, app.dh) })

	mock_token := uuid.NewString()
	request, e := http.NewRequest("GET", fmt.Sprintf("/confirm/%s", mock_token), nil)
	if e != nil {
		t.Fatal(e)
	}

	mock_id := uuid.NewString()
	app.database.ExpectQuery(`SELECT subscriber_id FROM subscription_tokens WHERE`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(
			pgxmock.NewRows([]string{"subscriber_id"}).
				AddRow(mock_id),
		)

	app.database.ExpectExec(`UPDATE subscriptions SET status = 'confirmed' WHERE`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	app.new_mock_request(request)
	defer app.database.ExpectationsWereMet()

	// tests
	if status := app.recorder.Code; status != http.StatusAccepted {
		t.Errorf("Expected status code %v, but got %v", http.StatusOK, status)
	}

	expected_body := `{"requestID":"","subscriber":"Subscription confirmed"}`
	response_body := app.recorder.Body.String()
	if response_body != expected_body {
		t.Errorf("Expected body %v, but got %v", expected_body, response_body)
	}
}

func Test_ConfirmSubscriber_InvalidID_Fails(t *testing.T) {
	// initialize
	app := new_mock_app()
	defer app.database.Close(app.context)

	app.router.GET("/confirm/:token", func(c *gin.Context) { routes.ConfirmSubscriber(c, app.dh) })

	mock_token := uuid.NewString()
	request, e := http.NewRequest("GET", fmt.Sprintf("/confirm/%s", mock_token), nil)
	if e != nil {
		t.Fatal(e)
	}

	invalid_token := uuid.NewString()
	mock_id := uuid.NewString()
	app.database.ExpectQuery(`SELECT subscriber_id FROM subscription_tokens WHERE`).
		WithArgs(pgxmock.AnyArg().Match(invalid_token)).
		WillReturnRows(
			pgxmock.NewRows([]string{"subscriber_id"}).
				AddRow(mock_id),
		)

	app.database.ExpectExec(`UPDATE subscriptions SET status = 'confirmed' WHERE`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	app.new_mock_request(request)
	defer app.database.ExpectationsWereMet()

	// tests
	if status := app.recorder.Code; status != http.StatusInternalServerError {
		t.Errorf("Expected status code %v, but got %v", http.StatusOK, status)
	}

	expected_body := fmt.Sprintf(`{"error":"Failed to fetch subscriber ID: argument 0 expected [bool - true] does not match actual [string - %s]","requestID":""}`, mock_token)
	response_body := app.recorder.Body.String()
	if response_body != expected_body {
		t.Errorf("Expected body %v, but got %v", expected_body, response_body)
	}
}
