package api_test

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/clients"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/handlers"
)

type App struct {
	recorder *httptest.ResponseRecorder
	context  *gin.Context
	router   *gin.Engine
	database pgxmock.PgxConnIface
	dh       *handlers.DatabaseHandler
	client   *clients.SMTPClient
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

var admin *gin.RouterGroup

func new_mock_app() App {
	var recorder *httptest.ResponseRecorder
	var context *gin.Context
	var database pgxmock.PgxConnIface
	var client *clients.SMTPClient
	var dh *handlers.DatabaseHandler
	var store cookie.Store

	recorder = httptest.NewRecorder()
	database = new_mock_database()
	client = new_mock_client()
	dh = handlers.NewDatabaseHandler(database)

	router := gin.Default()
	context = gin.CreateTestContextOnly(recorder, router)
	router.LoadHTMLGlob(filepath.Join("../api/templates", "*"))

	store = cookie.NewStore([]byte("test"))
	router.Use(sessions.Sessions("test", store))

	return App{
		recorder: recorder,
		context:  context,
		router:   router,
		database: database,
		dh:       dh,
		client:   client,
	}
}

func (app *App) new_mock_request(request *http.Request) {
	app.router.ServeHTTP(app.recorder, request)
}
