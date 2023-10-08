package api_test

import (
	"net/http"
	"net/http/httptest"

	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	mock "github.com/mocktools/go-smtp-mock"
	"github.com/pashagolub/pgxmock/v3"

	"github.com/solomonbaez/SB-Go-Newsletter-API/api/clients"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/handlers"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/models"
)

type app struct {
	recorder *httptest.ResponseRecorder
	context  *gin.Context
	router   *gin.Engine
}

func Test_HealthCheck_Returns_OK(t *testing.T) {
	// router
	router := gin.Default()
	router.GET("/health", handlers.HealthCheck)

	// server initialization
	request, e := http.NewRequest("GET", "/health", nil)
	if e != nil {
		t.Fatal(e)
	}

	// tests
	app := spawn_app(router, request)
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
	database, e := spawn_mock_database()
	if e != nil {
		t.Fatal(e)
	}
	client, e := spawn_mock_smtp_client()
	if e != nil {
		t.Fatal(e)
	}

	router := spawn_mock_router(database, client)

	request, e := http.NewRequest("GET", "/subscribers", nil)
	if e != nil {
		t.Fatal(e)
	}

	database.ExpectQuery(`SELECT \* FROM subscriptions`).WillReturnRows(
		pgxmock.NewRows([]string{"id", "email", "name", "created", "status"}),
	)

	app := spawn_app(router, request)
	defer database.ExpectationsWereMet()
	defer database.Close(app.context)

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
	database, e := spawn_mock_database()
	if e != nil {
		t.Fatal(e)
	}
	client, e := spawn_mock_smtp_client()
	if e != nil {
		t.Fatal(e)
	}

	router := spawn_mock_router(database, client)

	request, e := http.NewRequest("GET", "/subscribers", nil)
	if e != nil {
		t.Fatal(e)
	}

	mock_id := uuid.NewString()
	database.ExpectQuery(`SELECT \* FROM subscriptions`).WillReturnRows(
		pgxmock.NewRows([]string{"id", "email", "name", "created", "status"}).
			AddRow(mock_id, models.SubscriberEmail("test@test.com"), models.SubscriberName("TestUser"), time.Now(), "pending"),
	)

	app := spawn_app(router, request)
	defer database.ExpectationsWereMet()
	defer database.Close(app.context)

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
	database, e := spawn_mock_database()
	if e != nil {
		t.Fatal(e)
	}
	client, e := spawn_mock_smtp_client()
	if e != nil {
		t.Fatal(e)
	}

	router := spawn_mock_router(database, client)

	request, e := http.NewRequest("GET", "/confirmed", nil)
	if e != nil {
		t.Fatal(e)
	}

	database.ExpectQuery(`SELECT id, email, name, created, status FROM subscriptions WHERE`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(
			pgxmock.NewRows([]string{"id", "email", "name", "created", "status"}),
		)

	app := spawn_app(router, request)
	defer database.ExpectationsWereMet()
	defer database.Close(app.context)

	// tests
	if status := app.recorder.Code; status != http.StatusOK {
		t.Errorf("Expected status code %v, but got %v", http.StatusOK, status)
	}

	expected_body := `{"requestID":"","subscribers":"No confirmed subscribers"}`
	response_body := app.recorder.Body.String()
	if response_body != expected_body {
		t.Errorf("Expected body %v, but got %v", expected_body, response_body)
	}
}

func Test_GetConfirmedSubscribers_WithSubscribers_Passes(t *testing.T) {
	// initialize
	database, e := spawn_mock_database()
	if e != nil {
		t.Fatal(e)
	}
	client, e := spawn_mock_smtp_client()
	if e != nil {
		t.Fatal(e)
	}

	router := spawn_mock_router(database, client)

	request, e := http.NewRequest("GET", "/confirmed", nil)
	if e != nil {
		t.Fatal(e)
	}

	mock_id := uuid.NewString()
	database.ExpectQuery(`SELECT id, email, name, created, status FROM subscriptions WHERE`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(
			pgxmock.NewRows([]string{"id", "email", "name", "created", "status"}).
				AddRow(mock_id, models.SubscriberEmail("test@test.com"), models.SubscriberName("TestUser"), time.Now(), "confirmed"),
		)

	app := spawn_app(router, request)
	defer database.ExpectationsWereMet()
	defer database.Close(app.context)

	// tests
	if status := app.recorder.Code; status != http.StatusOK {
		t.Errorf("Expected status code %v, but got %v", http.StatusOK, status)
	}

	expected_body := fmt.Sprintf(`{"requestID":"","subscribers":[{"id":"%s","email":"test@test.com","name":"TestUser","status":"confirmed"}]}`, mock_id)
	response_body := app.recorder.Body.String()
	if response_body != expected_body {
		t.Errorf("Expected body %v, but got %v", expected_body, response_body)
	}
}

func Test_GetSubscribersByID_ValidID_Passes(t *testing.T) {
	// initialization
	database, e := spawn_mock_database()
	if e != nil {
		t.Fatal(e)
	}
	client, e := spawn_mock_smtp_client()
	if e != nil {
		t.Fatal(e)
	}

	router := spawn_mock_router(database, client)

	mock_id := uuid.NewString()

	request, e := http.NewRequest("GET", fmt.Sprintf("/subscribers/%v", mock_id), nil)
	if e != nil {
		t.Fatal(e)
	}

	database.ExpectQuery(`SELECT id, email, name, status FROM subscriptions WHERE`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(
			pgxmock.NewRows([]string{"id", "email", "name", "status"}).
				AddRow(mock_id, models.SubscriberEmail("test@test.com"), models.SubscriberName("TestUser"), "pending"),
		)

	// tests
	app := spawn_app(router, request)
	defer database.ExpectationsWereMet()
	defer database.Close(app.context)

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
	database, e := spawn_mock_database()
	if e != nil {
		t.Fatal(e)
	}
	client, e := spawn_mock_smtp_client()
	if e != nil {
		t.Fatal(e)
	}

	router := spawn_mock_router(database, client)

	// Non-UUID ID
	mock_id := "1"

	request, e := http.NewRequest("GET", fmt.Sprintf("/subscribers/%v", mock_id), nil)
	if e != nil {
		t.Fatal(e)
	}

	database.ExpectQuery(`SELECT id, email, name FROM subscriptions WHERE`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(
			pgxmock.NewRows([]string{"id", "email", "name"}).
				AddRow(mock_id, models.SubscriberEmail("test@test.com"), models.SubscriberName("TestUser")),
		)

	// tests
	app := spawn_app(router, request)
	defer database.ExpectationsWereMet()
	defer database.Close(app.context)

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
	database, e := spawn_mock_database()
	if e != nil {
		t.Fatal(e)
	}
	client, e := spawn_mock_smtp_client()
	if e != nil {
		t.Fatal(e)
	}

	router := spawn_mock_router(database, client)

	data := `{"email": "test@test.com", "name": "TestUser"}`
	request, e := http.NewRequest("POST", "/subscribe", strings.NewReader(data))
	if e != nil {
		t.Fatal(e)
	}

	database.ExpectBegin()
	database.ExpectExec("INSERT INTO subscriptions").
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	database.ExpectExec("INSERT INTO subscription_tokens").
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	database.ExpectCommit()

	app := spawn_app(router, request)
	defer database.ExpectationsWereMet()
	defer database.Close(app.context)

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
	var database pgxmock.PgxConnIface
	var client *clients.SMTPClient
	var router *gin.Engine
	var request *http.Request
	var app app
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
		database, e = spawn_mock_database()
		if e != nil {
			t.Fatal(e)
		}
		client, e = spawn_mock_smtp_client()
		if e != nil {
			t.Fatal(e)
		}

		router = spawn_mock_router(database, client)

		request, e = http.NewRequest("POST", "/subscribe", strings.NewReader(tc))
		if e != nil {
			t.Fatal(e)
		}

		database.ExpectBegin()
		database.ExpectExec("INSERT INTO subscriptions").
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg())
		database.ExpectRollback()

		app = spawn_app(router, request)
		defer database.ExpectationsWereMet()
		defer database.Close(app.context)

		// tests
		if status := app.recorder.Code; status != http.StatusBadRequest {
			t.Errorf("Expected status code %v, but got %v", http.StatusBadRequest, status)
		}
	}
}

func TestSubscribeInvalidNameFails(t *testing.T) {
	// // initialization
	var database pgxmock.PgxConnIface
	var client *clients.SMTPClient
	var router *gin.Engine
	var request *http.Request
	var app app
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
		database, e = spawn_mock_database()
		if e != nil {
			t.Fatal(e)
		}
		client, e = spawn_mock_smtp_client()
		if e != nil {
			t.Fatal(e)
		}

		router = spawn_mock_router(database, client)

		request, e = http.NewRequest("POST", "/subscribe", strings.NewReader(tc))
		if e != nil {
			t.Fatal(e)
		}

		database.ExpectBegin()
		database.ExpectExec("INSERT INTO subscriptions").
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg())
		database.ExpectRollback()

		app = spawn_app(router, request)
		defer database.ExpectationsWereMet()
		defer database.Close(app.context)

		// tests
		if status := app.recorder.Code; status != http.StatusBadRequest {
			t.Errorf("Expected status code %v, but got %v", http.StatusBadRequest, status)
		}
	}
}

func Test_Subscribe_MaxLengthParameters_Fails(t *testing.T) {
	// // initialization
	var database pgxmock.PgxConnIface
	var client *clients.SMTPClient
	var router *gin.Engine
	var request *http.Request
	var app app
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
		database, e = spawn_mock_database()
		if e != nil {
			t.Fatal(e)
		}
		client, e = spawn_mock_smtp_client()
		if e != nil {
			t.Fatal(e)
		}

		router = spawn_mock_router(database, client)

		request, e = http.NewRequest("POST", "/subscribe", strings.NewReader(tc))
		if e != nil {
			t.Fatal(e)
		}

		database.ExpectBegin()
		database.ExpectExec("INSERT INTO subscriptions").
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg())
		database.ExpectRollback()

		app = spawn_app(router, request)
		defer database.ExpectationsWereMet()
		defer database.Close(app.context)

		// tests
		if status := app.recorder.Code; status != http.StatusBadRequest {
			t.Errorf("Expected status code %v, but got %v", http.StatusBadRequest, status)
		}
	}
}

func Test_ConfirmSubscriber_Passes(t *testing.T) {
	// initialize
	database, e := spawn_mock_database()
	if e != nil {
		t.Fatal(e)
	}
	client, e := spawn_mock_smtp_client()
	if e != nil {
		t.Fatal(e)
	}

	router := spawn_mock_router(database, client)

	mock_token := uuid.NewString()
	request, e := http.NewRequest("GET", fmt.Sprintf("/confirm/%s", mock_token), nil)
	if e != nil {
		t.Fatal(e)
	}

	mock_id := uuid.NewString()
	database.ExpectQuery(`SELECT subscriber_id FROM subscription_tokens WHERE`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(
			pgxmock.NewRows([]string{"subscriber_id"}).
				AddRow(mock_id),
		)

	database.ExpectExec(`UPDATE subscriptions SET status = 'confirmed' WHERE`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	app := spawn_app(router, request)
	defer database.ExpectationsWereMet()
	defer database.Close(app.context)

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
	database, e := spawn_mock_database()
	if e != nil {
		t.Fatal(e)
	}
	client, e := spawn_mock_smtp_client()
	if e != nil {
		t.Fatal(e)
	}

	router := spawn_mock_router(database, client)

	mock_token := uuid.NewString()
	request, e := http.NewRequest("GET", fmt.Sprintf("/confirm/%s", mock_token), nil)
	if e != nil {
		t.Fatal(e)
	}

	invalid_token := uuid.NewString()
	mock_id := uuid.NewString()
	database.ExpectQuery(`SELECT subscriber_id FROM subscription_tokens WHERE`).
		WithArgs(pgxmock.AnyArg().Match(invalid_token)).
		WillReturnRows(
			pgxmock.NewRows([]string{"subscriber_id"}).
				AddRow(mock_id),
		)

	database.ExpectExec(`UPDATE subscriptions SET status = 'confirmed' WHERE`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	app := spawn_app(router, request)
	defer database.ExpectationsWereMet()
	defer database.Close(app.context)

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

func Test_PostNewsletter_Passes(t *testing.T) {
	// initialize
	database, e := spawn_mock_database()
	if e != nil {
		t.Fatal(e)
	}

	cfg := mock.ConfigurationAttr{}
	server := mock.New(cfg)
	server.Start()
	defer server.Stop()
	port := server.PortNumber

	client, e := spawn_mock_smtp_client()
	if e != nil {
		t.Fatal(e)
	}
	client.SmtpServer = "[::]"
	client.SmtpPort = port
	client.Sender = models.SubscriberEmail("user@test.com")

	body := models.Body{
		Title: "testing",
		Text:  "testing",
		Html:  "<p>testing</p>",
	}
	emailContent := models.Newsletter{
		Recipient: models.SubscriberEmail("recipient@test.com"),
		Content:   &body,
	}
	fmt.Printf(emailContent.Content.Html)

	router := spawn_mock_router(database, client)

	mock_username := "user"
	mock_password := "password"
	data := `{"title":"test", "text":"test", "html":"<p>test</p>"}`
	request, e := http.NewRequest("POST", "/newsletter", strings.NewReader(data))
	if e != nil {
		t.Fatal(e)
	}
	request.SetBasicAuth(mock_username, mock_password)

	mock_id := uuid.NewString()
	mock_password_hash, e := handlers.GeneratePHC(mock_password)
	if e != nil {
		t.Fatal(e)
	}
	database.ExpectQuery(`SELECT id, password_hash FROM users WHERE`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(
			pgxmock.NewRows([]string{"id", "password_hash"}).
				AddRow(mock_id, mock_password_hash),
		)

	database.ExpectQuery(`SELECT id, email, name, created, status FROM subscriptions WHERE`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(
			pgxmock.NewRows([]string{"id", "email", "name", "created", "status"}).
				AddRow(mock_id, models.SubscriberEmail("test@test.com"), models.SubscriberName("TestUser"), time.Now(), "confirmed"),
		)

	app := spawn_app(router, request)
	defer database.ExpectationsWereMet()
	defer database.Close(app.context)

	// tests
	if status := app.recorder.Code; status != http.StatusOK {
		t.Errorf("Expected status code %v, but got %v", http.StatusOK, status)
	}

	expected_body := fmt.Sprintf(
		`{"requestID":"","subscribers":[{"id":"%s","email":"test@test.com","name":"TestUser","status":"confirmed"}]}`, mock_id,
	) + `{"message":"Emails successfully delivered","requestID":""}`
	response_body := app.recorder.Body.String()
	if response_body != expected_body {
		t.Errorf("Expected body %v, but got %v", expected_body, response_body)
	}
}

func Test_PostNewsletter_InvalidPassword_Fails(t *testing.T) {
	// initialize
	database, e := spawn_mock_database()
	if e != nil {
		t.Fatal(e)
	}

	cfg := mock.ConfigurationAttr{}
	server := mock.New(cfg)
	server.Start()
	defer server.Stop()
	port := server.PortNumber

	client, e := spawn_mock_smtp_client()
	if e != nil {
		t.Fatal(e)
	}
	client.SmtpServer = "[::]"
	client.SmtpPort = port
	client.Sender = models.SubscriberEmail("user@test.com")

	body := models.Body{
		Title: "testing",
		Text:  "testing",
		Html:  "<p>testing</p>",
	}
	emailContent := models.Newsletter{
		Recipient: models.SubscriberEmail("recipient@test.com"),
		Content:   &body,
	}
	fmt.Printf(emailContent.Content.Html)

	router := spawn_mock_router(database, client)

	mock_username := "user"
	mock_password := "password"
	invalid_password := "drowssap"
	data := `{"title":"test", "text":"test", "html":"<p>test</p>"}`
	request, e := http.NewRequest("POST", "/newsletter", strings.NewReader(data))
	if e != nil {
		t.Fatal(e)
	}
	// I'm relatively unconcerned about basic auth failing in this integration test
	// TODO sketch out a unit test for handlers.BasicAuth
	request.SetBasicAuth(mock_username, mock_password)

	mock_id := uuid.NewString()
	invalid_password_hash, e := handlers.GeneratePHC(invalid_password)
	if e != nil {
		t.Fatal(e)
	}
	database.ExpectQuery(`SELECT id, password_hash FROM users WHERE`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(
			pgxmock.NewRows([]string{"id", "password_hash"}).
				AddRow(mock_id, invalid_password_hash),
		)

	app := spawn_app(router, request)
	defer database.ExpectationsWereMet()
	defer database.Close(app.context)

	// tests
	if status := app.recorder.Code; status != http.StatusBadRequest {
		t.Errorf("Expected status code %v, but got %v", http.StatusOK, status)
	}

	expected_body := `{"error":"Failed to validate credentials: PHC are not equivalent","requestID":""}`
	response_body := app.recorder.Body.String()
	if response_body != expected_body {
		t.Errorf("Expected body %v, but got %v", expected_body, response_body)
	}
}

func Test_PostNewsletter_InvalidNewsletter_Fails(t *testing.T) {
	// // initialization
	var database pgxmock.PgxConnIface
	var client *clients.SMTPClient
	var router *gin.Engine
	var request *http.Request
	var app app
	var e error

	cfg := mock.ConfigurationAttr{}
	server := mock.New(cfg)

	mock_username := "user"
	mock_password := "password"
	mock_id := uuid.NewString()

	test_cases := []*models.Body{
		{
			Title: "",
			Text:  "testing",
			Html:  "<p>testing</p>",
		},
		{
			Title: "testing",
			Text:  "",
			Html:  "<p>testing</p>",
		},
		{
			Title: "testing",
			Text:  "testing",
			Html:  "",
		},
	}
	expected_responses := []string{
		`{"error":"Failed to parse newsletter: field: Title cannot be empty","requestID":""}`,
		`{"error":"Failed to parse newsletter: field: Text cannot be empty","requestID":""}`,
		`{"error":"Failed to parse newsletter: field: Html cannot be empty","requestID":""}`,
	}

	for i, tc := range test_cases {
		// initialize
		database, e = spawn_mock_database()
		if e != nil {
			t.Fatal(e)
		}

		server.Start()
		defer server.Stop()
		port := server.PortNumber

		client, e = spawn_mock_smtp_client()
		if e != nil {
			t.Fatal(e)
		}
		client.SmtpServer = "[::]"
		client.SmtpPort = port
		client.Sender = models.SubscriberEmail("user@test.com")

		router = spawn_mock_router(database, client)

		mock_password_hash, e := handlers.GeneratePHC(mock_password)
		if e != nil {
			t.Fatal(e)
		}

		data := fmt.Sprintf(`{"title":"%s", "text":"%s", "html":"%s"}`, tc.Title, tc.Text, tc.Html)
		request, e = http.NewRequest("POST", "/newsletter", strings.NewReader(data))
		if e != nil {
			t.Fatal(e)
		}
		request.SetBasicAuth(mock_username, mock_password)

		database.ExpectQuery(`SELECT id, password_hash FROM users WHERE`).
			WithArgs(pgxmock.AnyArg()).
			WillReturnRows(
				pgxmock.NewRows([]string{"id", "password_hash"}).
					AddRow(mock_id, mock_password_hash),
			)

		database.ExpectQuery(`SELECT id, email, name, created, status FROM subscriptions WHERE`).
			WithArgs(pgxmock.AnyArg()).
			WillReturnRows(
				pgxmock.NewRows([]string{"id", "email", "name", "created", "status"}).
					AddRow(mock_id, models.SubscriberEmail("test@test.com"), models.SubscriberName("TestUser"), time.Now(), "confirmed"),
			)

		app = spawn_app(router, request)
		defer database.ExpectationsWereMet()
		defer database.Close(app.context)

		// tests
		if status := app.recorder.Code; status != http.StatusBadRequest {
			t.Errorf("Expected status code %v, but got %v", http.StatusOK, status)
		}

		response_body := app.recorder.Body.String()
		if response_body != expected_responses[i] {
			t.Errorf("Expected body %v, but got %v", expected_responses[i], response_body)
		}
	}
}

func spawn_mock_database() (pgxmock.PgxConnIface, error) {
	mock_db, e := pgxmock.NewConn()
	if e != nil {
		return nil, e
	}

	return mock_db, nil
}

func spawn_mock_smtp_client() (*clients.SMTPClient, error) {
	cfg := "test"
	client, e := clients.NewSMTPClient(&cfg)
	if e != nil {
		return nil, e
	}

	return client, nil
}

func spawn_mock_router(db pgxmock.PgxConnIface, client *clients.SMTPClient) *gin.Engine {
	rh := handlers.NewRouteHandler(db)

	router := gin.Default()
	router.GET("/subscribers", rh.GetSubscribers)
	router.GET("/confirmed", func(c *gin.Context) {
		_ = rh.GetConfirmedSubscribers(c)
	})
	router.GET("/confirm/:token", rh.ConfirmSubscriber)
	router.GET("/subscribers/:id", rh.GetSubscriberByID)
	router.POST("/subscribe", func(c *gin.Context) {
		rh.Subscribe(c, client)
	})
	router.POST("/newsletter", func(c *gin.Context) {
		rh.PostNewsletter(c, client)
	})

	return router
}

func spawn_app(router *gin.Engine, request *http.Request) app {
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	context, _ := gin.CreateTestContext(recorder)

	return app{
		recorder,
		context,
		router,
	}
}
