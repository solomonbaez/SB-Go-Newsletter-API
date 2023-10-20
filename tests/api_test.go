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

	"github.com/solomonbaez/SB-Go-Newsletter-API/api/handlers"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/models"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/routes"
	adminRoutes "github.com/solomonbaez/SB-Go-Newsletter-API/api/routes/admin"
)

func TestAPI(t *testing.T) {
	tests := []struct {
		name string
		fn   func(*testing.T)
	}{
		{"HealthCheck", healthCheck},
		{"GetSubscribers", getSubscribers},
		{"GetConfirmedSubscribers", getConfirmedSubscribers},
		{"GetSubscriberByID", getSubscribersByID},
	}

	for _, test := range tests {
		t.Run(test.name, test.fn)
	}
}

func healthCheck(t *testing.T) {
	test := &struct {
		name           string
		expectedStatus int
	}{
		"(+) Test case -> GET request to /health -> passes",
		http.StatusOK,
	}

	t.Parallel()
	// initialize
	app := new_mock_app()
	app.router.GET("/health", handlers.HealthCheck)
	defer app.database.Close(app.context)

	request, e := http.NewRequest("GET", "/health", nil)
	if e != nil {
		t.Fatal(e)
	}

	// assertions
	app.new_mock_request(request)
	if responseStatus := app.recorder.Code; responseStatus != test.expectedStatus {
		t.Errorf("Expected status code %v, but got %v", http.StatusOK, responseStatus)
	}

	expected_body := `"OK"`
	response_body := app.recorder.Body.String()
	if response_body != expected_body {
		t.Errorf("Expected body %v, but got %v", expected_body, response_body)
	}
}

func getSubscribers(t *testing.T) {
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
			`{"requestID":"","subscribers":"No subscribers"}`,
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
		app := new_mock_app()
		admin = app.router.Group("/admin")
		admin.GET("/subscribers", func(c *gin.Context) { adminRoutes.GetSubscribers(c, app.dh) })
		defer app.database.Close(app.context)

		request, e := http.NewRequest("GET", "/admin/subscribers", nil)
		if e != nil {
			t.Fatal(e)
		}

		if tc.subscribers {
			app.database.ExpectQuery(`SELECT \* FROM subscriptions`).
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
			app.database.ExpectQuery(`SELECT \* FROM subscriptions`).
				WillReturnRows(
					pgxmock.NewRows([]string{"id", "email", "name", "created", "status"}),
				)
		}

		app.new_mock_request(request)
		defer app.database.ExpectationsWereMet()

		// tests
		if responseStatus := app.recorder.Code; responseStatus != tc.expectedStatus {
			t.Errorf("Expected status code %v, but got %v", tc.expectedStatus, responseStatus)
		}

		responseBody := app.recorder.Body.String()
		if responseBody != tc.expectedBody {
			t.Errorf("Expected body %v, but got %v", tc.expectedBody, responseBody)
		}
	}
}

func getConfirmedSubscribers(t *testing.T) {
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
	app := new_mock_app()
	defer app.database.Close(app.context)

	app.database.ExpectQuery(`SELECT id, email, name, created, status FROM subscriptions WHERE`).
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

	responseArray := adminRoutes.GetConfirmedSubscribers(app.context, app.dh)
	defer app.database.ExpectationsWereMet()

	if *responseArray[0] != *test.expectedArray[0] {
		t.Errorf("Expected array: %v, got: %v", *test.expectedArray[0], *responseArray[0])
	}
}

func getSubscribersByID(t *testing.T) {
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
		app := new_mock_app()
		admin = app.router.Group("/admin")
		admin.GET("/subscribers/:id", func(c *gin.Context) { adminRoutes.GetSubscriberByID(c, app.dh) })
		defer app.database.Close(app.context)

		request, e := http.NewRequest("GET", fmt.Sprintf("/admin/subscribers/%v", seedSubscriber.id), nil)
		if e != nil {
			t.Fatal(e)
		}

		query := app.database.ExpectQuery(`SELECT id, email, name, status FROM subscriptions WHERE`).
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
		app.new_mock_request(request)
		defer app.database.ExpectationsWereMet()

		if responseStatus := app.recorder.Code; responseStatus != tc.expectedStatus {
			t.Errorf("Expected status code %v, but got %v", tc.expectedStatus, responseStatus)
		}

		responseBody := app.recorder.Body.String()
		if responseBody != tc.expectedBody {
			t.Errorf("Expected body %v, but got %v", tc.expectedBody, responseBody)
		}
	}
}

// TODO clean up routes.Subscribe + routes.insertSubscriber
func Test_Subscribe_Passes(t *testing.T) {
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
		data           string
		expectedStatus int
		expectedBody   string
	}{
		{
			"(+) Test case -> POST to /subscribe with valid fields -> passes",
			fmt.Sprintf(
				`{"email": "%s", "name": "%s"}`,
				seedSubscriber.email.String(),
				seedSubscriber.name.String(),
			),

			http.StatusCreated,

			fmt.Sprintf(
				`{"requestID":"","subscriber":{"id":"","email":"%s","name":"%s","status":"%s"}}`,
				seedSubscriber.email,
				seedSubscriber.name,
				seedSubscriber.status,
			),
		},
		{
			"(-) Test case -> POST to /subscribe with invalid email -> fails",
			fmt.Sprintf(
				`{"email": "%s", "name": "%s"}`,
				"user",
				seedSubscriber.name.String(),
			),

			http.StatusBadRequest,

			`{"error":"Could not subscribe: invalid email format","requestID":""}`,
		},
	}

	t.Parallel()
	for _, tc := range *testCases {
		// initialization
		app := new_mock_app()
		app.router.POST("/subscribe", func(c *gin.Context) { routes.Subscribe(c, app.dh, app.client) })
		defer app.database.Close(app.context)

		request, _ := http.NewRequest("POST", "/subscribe", strings.NewReader(tc.data))

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
		if responseStatus := app.recorder.Code; responseStatus != tc.expectedStatus {
			t.Errorf("Expected status code %v, but got %v", tc.expectedStatus, responseStatus)
		}

		responseBody := app.recorder.Body.String()
		if responseBody != tc.expectedBody {
			t.Errorf("Expected body %v, but got %v", tc.expectedBody, responseBody)
		}
	}
}

func Test_Subscribe_InvalidEmail_Fails(t *testing.T) {
	// // initialization
	var app App
	var request *http.Request
	var e error

	var test_cases []string
	test_cases = append(test_cases)
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
