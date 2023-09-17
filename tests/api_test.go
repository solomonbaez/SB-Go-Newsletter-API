package api_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/solomonbaez/SB-Go-Newsletter-API/api/handlers"
)

var router *gin.Engine

func init() {
	router = gin.Default()
	router.GET("/health", handlers.HealthCheck)
	router.POST("/subscribe", handlers.Subscribe)
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

func TestGetSubscribersNoSubscribers(t *testing.T) {
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

func TestGetSubscribersWithSubscribers(t *testing.T) {
	// server initialization
	data := `{"email": "test@test.com", "name": "test"}`
	post_request, e := http.NewRequest("POST", "/subscribe", strings.NewReader(data))
	if e != nil {
		t.Fatal(e)
	}

	app := spawn_app(post_request)

	//  test POST
	if status := app.recorder.Code; status != http.StatusCreated {
		t.Errorf("Expected status code %v, but got %v", http.StatusCreated, status)
	}

	// test GET
	t.Run("GetSubscribers", func(t *testing.T) {
		get_request, e := http.NewRequest("GET", "/subscribers", nil)
		if e != nil {
			t.Fatal(e)
		}

		app.recorder = httptest.NewRecorder()
		app.router.ServeHTTP(app.recorder, get_request)

		// tests
		if status := app.recorder.Code; status != http.StatusOK {
			t.Errorf("Expected status code %v, but got %v", http.StatusOK, status)
		}

		expected_body := `{"test@test.com":{"email":"test@test.com","name":"test"}}`
		response_body := app.recorder.Body.String()

		if response_body != expected_body {
			t.Errorf("Expected body %v, but got %v", expected_body, response_body)
		}
	})
}
