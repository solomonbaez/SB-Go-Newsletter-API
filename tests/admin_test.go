package api_test

import (
	"net/http"
	"net/http/httptest"
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

// Reference
type App struct {
	recorder *httptest.ResponseRecorder
	context  *gin.Context
	router   *gin.Engine
	database pgxmock.PgxConnIface
	client   *clients.SMTPClient
}

func Test_GetLogin(t *testing.T) {
	// initialize
	app := new_mock_app()
	defer app.database.Close(app.context)

	request, e := http.NewRequest("GET", "/admin/login", nil)
	if e != nil {
		t.Fatal(e)
	}

	app.new_mock_request(request)

	// tests
	if status := app.recorder.Code; status != http.StatusOK {
		t.Errorf("Expected status code %v, but got %v", http.StatusOK, status)
	}
}

func Test_PostLogin_Passes(t *testing.T) {
	// initialize
	app := new_mock_app()
	defer app.database.Close(app.context)

	request, e := http.NewRequest("POST", "/admin/login", nil)
	if e != nil {
		t.Fatal(e)
	}

	app.database.ExpectQuery(`SELECT id, password_hash FROM users WHERE user = $1`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(
			pgxmock.NewRows([]string{"id", "password_hash"}),
		)

	app.new_mock_request(request)

	// tests
	if status := app.recorder.Code; status != http.StatusSeeOther {
		t.Errorf("Expected status code %v, but got %v", http.StatusOK, status)
	}
}

func new_mock_database() (database pgxmock.PgxConnIface) {
	database, _ = pgxmock.NewConn()

	return database
}

func new_mock_client() (client *clients.SMTPClient) {
	cfg := "test"
	client, _ = clients.NewSMTPClient(&cfg)

	return client
}

func new_mock_app() (app App) {
	database := new_mock_database()
	client := new_mock_client()
	rh := handlers.NewRouteHandler(database)

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

	app = App{
		router:   router,
		database: database,
		client:   client,
	}

	return app
}

func (app *App) new_mock_request(request *http.Request) {
	recorder := httptest.NewRecorder()
	app.recorder = recorder
	app.router.ServeHTTP(recorder, request)

	context, _ := gin.CreateTestContext(recorder)
	app.context = context
}
