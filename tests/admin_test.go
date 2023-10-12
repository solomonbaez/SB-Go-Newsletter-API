package api_test

import (
	"net/http"

	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/clients"
)

// // Reference
// type app struct {
// 	recorder *httptest.ResponseRecorder
// 	context  *gin.Context
// 	router   *gin.Engine
// }

func TestGetLogin(t *testing.T) {
	// initialize
	database, e := spawn_mock_database()
	if e != nil {
		t.Fatal(e)
	}
	client, e := spawn_mock_smtp_client()
	if e != nil {
		t.Fatal(e)
	}

	router := admin_spawn_mock_router(database, client)

	request, e := http.NewRequest("GET", "/admin/login", nil)
	if e != nil {
		t.Fatal(e)
	}

	app := spawn_app(router, request)
	defer database.ExpectationsWereMet()
	defer database.Close(app.context)

	// tests
	if status := app.recorder.Code; status != http.StatusOK {
		t.Errorf("Expected status code %v, but got %v", http.StatusOK, status)
	}

	// expected_body := `{"requestID":"","subscribers":"No subscribers"}`
	// response_body := app.recorder.Body.String()
	// if response_body != expected_body {
	// 	t.Errorf("Expected body %v, but got %v", expected_body, response_body)
	// }
}

// // Reference
// func spawn_mock_database() (pgxmock.PgxConnIface, error) {
// 	mock_db, e := pgxmock.NewConn()
// 	if e != nil {
// 		return nil, e
// 	}

// 	return mock_db, nil
// }

// func spawn_mock_smtp_client() (*clients.SMTPClient, error) {
// 	cfg := "test"
// 	client, e := clients.NewSMTPClient(&cfg)
// 	if e != nil {
// 		return nil, e
// 	}

// 	return client, nil
// }

func admin_spawn_mock_router(db pgxmock.PgxConnIface, client *clients.SMTPClient) *gin.Engine {
	// rh := handlers.NewRouteHandler(db)

	router := gin.Default()
	admin := router.Group("/admin")
	admin.GET("/login")

	return router
}

// func spawn_app(router *gin.Engine, request *http.Request) app {
// 	recorder := httptest.NewRecorder()
// 	router.ServeHTTP(recorder, request)
// 	context, _ := gin.CreateTestContext(recorder)

// 	return app{
// 		recorder,
// 		context,
// 		router,
// 	}
// }
