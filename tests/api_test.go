package api_test

import (
	"net/http"
	"net/http/httptest"

	"fmt"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/handlers"
)

type App struct {
	recorder *httptest.ResponseRecorder
	router   *gin.Engine
}

func spawn_app(router *gin.Engine, request *http.Request) App {
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	return App{
		recorder,
		router,
	}
}

func TestHealthCheckReturnsOK(t *testing.T) {
	// router
	router := gin.Default()
	router.GET("/health", handlers.HealthCheck)

	// server initialization
	request, e := http.NewRequest("GET", "/health", nil)
	if e != nil {
		t.Fatal(e)
	}

	app := spawn_app(router, request)

	// tests
	if status := app.recorder.Code; status != http.StatusOK {
		t.Errorf("Expected status code %v, but got %v", http.StatusOK, status)
	}

	expected_body := `"OK"`
	response_body := app.recorder.Body.String()

	if response_body != expected_body {
		t.Errorf("Expected body %v, but got %v", expected_body, response_body)
	}
}

func TestGetSubscribersNoSubscribers(t *testing.T) {
	db, e := SpawnMockDatabase()
	if e != nil {
		t.Fatal(e)
	}
	defer db.ExpectationsWereMet()

	router := SpawnMockRouter(db)

	// server initialization
	request, e := http.NewRequest("GET", "/subscribers", nil)
	if e != nil {
		t.Fatal(e)
	}

	db.ExpectQuery(`SELECT \* FROM subscriptions`).WillReturnRows(
		pgxmock.NewRows([]string{"id", "email", "name", "created"}),
	)
	app := spawn_app(router, request)

	// tests
	if status := app.recorder.Code; status != http.StatusOK {
		t.Errorf("Expected status code %v, but got %v", http.StatusOK, status)
	}

	expected_body := `"No subscribers"`
	response_body := app.recorder.Body.String()

	if response_body != expected_body {
		t.Errorf("Expected body %v, but got %v", expected_body, response_body)
	}
}

func TestGetSubscribersWithSubscribers(t *testing.T) {
	db, e := SpawnMockDatabase()
	if e != nil {
		t.Fatal(e)
	}
	defer db.ExpectationsWereMet()

	router := SpawnMockRouter(db)

	mock_id := uuid.NewString()

	request, e := http.NewRequest("GET", "/subscribers", nil)
	if e != nil {
		t.Fatal(e)
	}

	// Set up expectation for the SELECT query
	db.ExpectQuery(`SELECT \* FROM subscriptions`).WillReturnRows(
		pgxmock.NewRows([]string{"id", "email", "name", "created"}).
			AddRow(mock_id, "test@test.com", "Test User", time.Now()),
	)

	app := spawn_app(router, request)

	if status := app.recorder.Code; status != http.StatusOK {
		t.Errorf("Expected status code %v, but got %v", http.StatusOK, status)
	}

	expectedBody := fmt.Sprintf(`[{"id":"%v","email":"test@test.com","name":"Test User"}]`, mock_id)
	responseBody := app.recorder.Body.String()

	if responseBody != expectedBody {
		t.Errorf("Expected body %v, but got %v", expectedBody, responseBody)
	}
}

func TestGetSubscribersByID(t *testing.T) {
	db, e := SpawnMockDatabase()
	if e != nil {
		t.Fatal(e)
	}
	defer db.ExpectationsWereMet()

	router := SpawnMockRouter(db)

	mock_id := uuid.NewString()

	request, e := http.NewRequest("GET", fmt.Sprintf("/subscribers/%v", mock_id), nil)
	if e != nil {
		t.Fatal(e)
	}

	// Set up expectation for the SELECT query
	db.ExpectQuery(fmt.Sprintf(`SELECT id, email, name FROM subscriptions WHERE id=%v`, mock_id)).WillReturnRows(
		pgxmock.NewRows([]string{"id", "email", "name"}).
			AddRow(mock_id, "test@test.com", "Test User"),
	)

	app := spawn_app(router, request)

	if status := app.recorder.Code; status != http.StatusFound {
		t.Errorf("Expected status code %v, but got %v", http.StatusFound, status)
	}

	expectedBody := fmt.Sprintf(`{"id":"%v","email":"test@test.com","name":"Test User"}`, mock_id)
	responseBody := app.recorder.Body.String()

	if responseBody != expectedBody {
		t.Errorf("Expected body %v, but got %v", expectedBody, responseBody)
	}
}

func SpawnMockDatabase() (pgxmock.PgxConnIface, error) {
	mock_db, e := pgxmock.NewConn()
	if e != nil {
		return nil, e
	}

	return mock_db, nil
}

func SpawnMockRouter(db pgxmock.PgxConnIface) *gin.Engine {
	rh := handlers.NewRouteHandler(db)

	router := gin.Default()
	router.POST("/subscribe", rh.Subscribe)
	router.GET("/subscribers", rh.GetSubscribers)
	router.GET("/subscribers/:id", rh.GetSubscriberByID)

	return router
}

// // POSSIBLE POST TEST
// db.ExpectExec(`INSERT INTO subscriptions`).WillReturnResult(pgxmock.NewResult("INSERT", 1))

// // server initialization
// data := `{"email": "test@test.com", "name": "Test User"}`
// postRequest, e := http.NewRequest("POST", "/subscribe", strings.NewReader(data))
// if e != nil {
// 	t.Fatal(e)
// }

// app := spawn_app(postRequest)

// if status := app.recorder.Code; status != http.StatusCreated {
// 	t.Errorf("Expected status code %v, but got %v", http.StatusCreated, status)
// }
