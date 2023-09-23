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
	"github.com/pashagolub/pgxmock/v3"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/handlers"
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
	db, e := spawn_mock_database()
	if e != nil {
		t.Fatal(e)
	}
	router := spawn_mock_router(db)

	request, e := http.NewRequest("GET", "/subscribers", nil)
	if e != nil {
		t.Fatal(e)
	}

	db.ExpectQuery(`SELECT \* FROM subscriptions`).WillReturnRows(
		pgxmock.NewRows([]string{"id", "email", "name", "created"}),
	)

	app := spawn_app(router, request)
	defer db.ExpectationsWereMet()
	defer db.Close(app.context)

	// tests
	if status := app.recorder.Code; status != http.StatusOK {
		t.Errorf("Expected status code %v, but got %v", http.StatusOK, status)
	}

	expected_body := `{"request_id":"","subscribers":"No subscribers"}`
	response_body := app.recorder.Body.String()
	if response_body != expected_body {
		t.Errorf("Expected body %v, but got %v", expected_body, response_body)
	}
}

func Test_GetSubscribers_WithSubscribers_Passes(t *testing.T) {
	// initialization
	db, e := spawn_mock_database()
	if e != nil {
		t.Fatal(e)
	}
	router := spawn_mock_router(db)

	request, e := http.NewRequest("GET", "/subscribers", nil)
	if e != nil {
		t.Fatal(e)
	}

	mock_id := uuid.NewString()
	db.ExpectQuery(`SELECT \* FROM subscriptions`).WillReturnRows(
		pgxmock.NewRows([]string{"id", "email", "name", "created"}).
			AddRow(mock_id, "test@test.com", "Test User", time.Now()),
	)

	app := spawn_app(router, request)
	defer db.ExpectationsWereMet()
	defer db.Close(app.context)

	// tests
	if status := app.recorder.Code; status != http.StatusOK {
		t.Errorf("Expected status code %v, but got %v", http.StatusOK, status)
	}

	expected_body := fmt.Sprintf(`{"request_id":"","subscribers":[{"id":"%v","email":"test@test.com","name":"Test User"}]}`, mock_id)
	response_body := app.recorder.Body.String()
	if response_body != expected_body {
		t.Errorf("Expected body %v, but got %v", expected_body, response_body)
	}
}

func Test_GetSubscribersByID_ValidID_Passes(t *testing.T) {
	// initialization
	db, e := spawn_mock_database()
	if e != nil {
		t.Fatal(e)
	}
	router := spawn_mock_router(db)

	mock_id := uuid.NewString()

	request, e := http.NewRequest("GET", fmt.Sprintf("/subscribers/%v", mock_id), nil)
	if e != nil {
		t.Fatal(e)
	}

	db.ExpectQuery(`SELECT id, email, name FROM subscriptions WHERE`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(
			pgxmock.NewRows([]string{"id", "email", "name"}).
				AddRow(mock_id, "test@test.com", "Test User"),
		)

	// tests
	app := spawn_app(router, request)
	defer db.ExpectationsWereMet()
	defer db.Close(app.context)

	if status := app.recorder.Code; status != http.StatusFound {
		t.Errorf("Expected status code %v, but got %v", http.StatusFound, status)
	}

	expected_body := fmt.Sprintf(`{"request_id":"","subscriber":{"id":"%v","email":"test@test.com","name":"Test User"}}`, mock_id)
	response_body := app.recorder.Body.String()

	if response_body != expected_body {
		t.Errorf("Expected body %v, but got %v", expected_body, response_body)
	}
}

func Test_GetSubscribersByID_InvalidID_Fails(t *testing.T) {
	// initialization
	db, e := spawn_mock_database()
	if e != nil {
		t.Fatal(e)
	}
	router := spawn_mock_router(db)

	// Non-UUID ID
	mock_id := "1"

	request, e := http.NewRequest("GET", fmt.Sprintf("/subscribers/%v", mock_id), nil)
	if e != nil {
		t.Fatal(e)
	}

	db.ExpectQuery(`SELECT id, email, name FROM subscriptions WHERE`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(
			pgxmock.NewRows([]string{"id", "email", "name"}).
				AddRow(mock_id, "test@test.com", "Test User"),
		)

	// tests
	app := spawn_app(router, request)
	defer db.ExpectationsWereMet()
	defer db.Close(app.context)

	if status := app.recorder.Code; status != http.StatusBadRequest {
		t.Errorf("Expected status code %v, but got %v", http.StatusBadRequest, status)
	}

	expected_body := `{"error":"Invalid ID format, invalid UUID length: 1","request_id":""}`
	response_body := app.recorder.Body.String()

	if response_body != expected_body {
		t.Errorf("Expected body %v, but got %v", expected_body, response_body)
	}
}

func Test_Subscribe(t *testing.T) {
	// initialization
	db, e := spawn_mock_database()
	if e != nil {
		t.Fatal(e)
	}
	router := spawn_mock_router(db)

	data := `{"email": "test@test.com", "name": "Test User"}`
	request, e := http.NewRequest("POST", "/subscribe", strings.NewReader(data))
	if e != nil {
		t.Fatal(e)
	}

	db.ExpectExec("INSERT INTO subscriptions").
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	app := spawn_app(router, request)
	defer db.ExpectationsWereMet()
	defer db.Close(app.context)

	// tests
	if status := app.recorder.Code; status != http.StatusCreated {
		t.Errorf("Expected status code %v, but got %v", http.StatusCreated, status)
	}

	expected_body := `{"request_id":"","subscriber":{"id":"","email":"test@test.com","name":"Test User"}}`
	response_body := app.recorder.Body.String()
	if response_body != expected_body {
		t.Errorf("Expected body %v, but got %v", expected_body, response_body)
	}
}

func Test_Subscribe_InvalidEmail_Fails(t *testing.T) {
	// initialization
	db, e := spawn_mock_database()
	if e != nil {
		t.Fatal(e)
	}
	router := spawn_mock_router(db)

	var test_cases []string
	test_cases = append(test_cases,
		`{"email": "test", "name": "Test User"}`,
		`{"email": "test@", "name": "Test User"}`,
		`{"email": "@test.com", "name": "Test User"}`,
		`{"email": "test.com", "name": "Test User"}`,
	)
	for _, tc := range test_cases {
		request, e := http.NewRequest("POST", "/subscribe", strings.NewReader(tc))
		if e != nil {
			t.Fatal(e)
		}

		db.ExpectExec("INSERT INTO subscriptions").
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg())
			// WillReturnResult(pgxmock.NewResult("INSERT", 1))

		app := spawn_app(router, request)
		defer db.ExpectationsWereMet()
		defer db.Close(app.context)

		// tests
		if status := app.recorder.Code; status != http.StatusBadRequest {
			t.Errorf("Expected status code %v, but got %v", http.StatusBadRequest, status)
		}

		expected_body := `{"error":"invalid email format","request_id":""}`
		response_body := app.recorder.Body.String()
		if response_body != expected_body {
			t.Errorf("Expected body %v, but got %v", expected_body, response_body)
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

func spawn_mock_router(db pgxmock.PgxConnIface) *gin.Engine {
	rh := handlers.NewRouteHandler(db)

	router := gin.Default()
	router.POST("/subscribe", rh.Subscribe)
	router.GET("/subscribers", rh.GetSubscribers)
	router.GET("/subscribers/:id", rh.GetSubscriberByID)

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
