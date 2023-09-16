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
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, status)
	}

	expectedBody := `"OK"`
	if app.recorder.Body.String() != expectedBody {
		t.Errorf("Expected body %s, but got %s", expectedBody, app.recorder.Body.String())
	}
}
