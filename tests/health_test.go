package health_test

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/solomonbaez/SB-Go-Newsletter-API/api/handlers"
)

func TestHealthCheck(t *testing.T) {
	router := gin.Default()

	router.GET("/health", handlers.HealthCheck)

	recorder := httptest.NewRecorder()
	request, e := http.NewRequest("GET", "/health", nil)
	if e != nil {
		t.Fatal(e)
	}

	router.ServeHTTP(recorder, request)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, status)
	}

	expectedBody := `"OK"`
	if recorder.Body.String() != expectedBody {
		t.Errorf("Expected body %s, but got %s", expectedBody, recorder.Body.String())
	}
}
