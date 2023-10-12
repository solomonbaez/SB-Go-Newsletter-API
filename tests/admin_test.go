package api_test

import (
	"net/http"
	"path/filepath"

	"testing"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/clients"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/handlers"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/routes"
)

// // Reference
// type app struct {
// 	recorder *httptest.ResponseRecorder
// 	context  *gin.Context
// 	router   *gin.Engine
// }

func Test_GetLogin(t *testing.T) {
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
}

func Test_PostLogin_Passes(t *testing.T) {
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

	request, e := http.NewRequest("POST", "/admin/login", nil)
	if e != nil {
		t.Fatal(e)
	}

	database.ExpectQuery(`SELECT id, password_hash FROM users WHERE user = $1`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(
			pgxmock.NewRows([]string{"id", "password_hash"}),
		)

	app := spawn_app(router, request)
	defer database.ExpectationsWereMet()
	defer database.Close(app.context)

	// tests
	if status := app.recorder.Code; status != http.StatusSeeOther {
		t.Errorf("Expected status code %v, but got %v", http.StatusOK, status)
	}
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
	rh := handlers.NewRouteHandler(db)

	key1, _ := handlers.GenerateCSPRNG(32)
	key2, _ := handlers.GenerateCSPRNG(32)
	storeKeys := [][]byte{
		[]byte(key1),
		[]byte(key2),
	}
	store := cookie.NewStore(storeKeys...)

	router := gin.Default()
	router.LoadHTMLGlob(filepath.Join("../api/templates", "*"))

	router.Use(sessions.Sessions("test", store))

	admin := router.Group("/admin")
	admin.GET("/login", routes.GetLogin)
	admin.POST("/login", func(c *gin.Context) {
		routes.PostLogin(c, rh)
	})

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
