package api_test

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/solomonbaez/SB-Go-Newsletter-API/api/handlers"
)

var router *gin.Engine

func init() {
	router = gin.Default()
	router.GET("/health", handlers.HealthCheck)
	router.GET("/subscribers", handlers.GetSubscribers)
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

func TestGetSubscribers(t *testing.T) {
	// server initialization
	request, e := http.NewRequest("GET", "/subscribers", nil)
	if e != nil {
		t.Fatal(e)
	}

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
