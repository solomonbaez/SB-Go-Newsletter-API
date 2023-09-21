package api_test

import (
	"net/http"
	"net/http/httptest"

	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/handlers"
)

var router *gin.Engine
var c *gin.Context

func init() {
	router = gin.Default()
	router.GET("/health", handlers.HealthCheck)
}

type App struct {
	recorder *httptest.ResponseRecorder
	router   *gin.Engine
}

func spawn_app(request *http.Request) App {
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	return App{
		recorder,
		router,
	}
}

func TestHealthCheckReturnsOK(t *testing.T) {
	// server initialization
	request, e := http.NewRequest("GET", "/health", nil)
	if e != nil {
		t.Fatal(e)
	}

	app := spawn_app(request)

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

	// server initialization
	request, e := http.NewRequest("GET", "/subscribers", nil)
	if e != nil {
		t.Fatal(e)
	}

	db.ExpectQuery(`SELECT \* FROM subscriptions`).WillReturnRows(
		pgxmock.NewRows([]string{"id", "email", "name", "created"}),
	)
	app := spawn_app(request)

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

	request, e := http.NewRequest("GET", "/subscribers", nil)
	if e != nil {
		t.Fatal(e)
	}

	// Set up expectation for the SELECT query
	db.ExpectQuery(`SELECT \* FROM subscriptions`).WillReturnRows(
		pgxmock.NewRows([]string{"id", "email", "name", "created"}).
			AddRow("1", "test@test.com", "Test User", time.Now()),
	)

	app := spawn_app(request)

	if status := app.recorder.Code; status != http.StatusOK {
		t.Errorf("Expected status code %v, but got %v", http.StatusOK, status)
	}

	expectedBody := `[{"id":"1","email":"test@test.com","name":"Test User"}]`
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
	MockRouter(mock_db)
	return mock_db, nil
}

func MockRouter(db pgxmock.PgxConnIface) {
	rh := handlers.NewRouteHandler(db)
	router.POST("/subscribe", rh.Subscribe)
	router.GET("/subscribers", rh.GetSubscribers)
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
