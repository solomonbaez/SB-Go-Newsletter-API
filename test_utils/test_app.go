package utils

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/solomonbaez/hyacinth/api/clients"
	"github.com/solomonbaez/hyacinth/api/handlers"
)

type App struct {
	Recorder *httptest.ResponseRecorder
	Context  *gin.Context
	Router   *gin.Engine
	Database pgxmock.PgxConnIface
	DH       *handlers.DatabaseHandler
	Client   *clients.SMTPClient
}

func newMockDatabase() (database pgxmock.PgxConnIface) {
	database, _ = pgxmock.NewConn()

	return database
}

func newMockClient() (client *clients.SMTPClient) {
	cfg := "test"
	client, _ = clients.NewSMTPClient(&cfg)

	return client
}

func NewMockApp() App {
	var recorder *httptest.ResponseRecorder
	var context *gin.Context
	var database pgxmock.PgxConnIface
	var client *clients.SMTPClient
	var dh *handlers.DatabaseHandler
	var store cookie.Store

	recorder = httptest.NewRecorder()
	database = newMockDatabase()
	client = newMockClient()
	dh = handlers.NewDatabaseHandler(database)

	router := gin.Default()
	context = gin.CreateTestContextOnly(recorder, router)
	router.LoadHTMLGlob(filepath.Join("../api/templates", "*"))

	store = cookie.NewStore([]byte("test"))
	router.Use(sessions.Sessions("test", store))

	return App{
		Recorder: recorder,
		Context:  context,
		Router:   router,
		Database: database,
		DH:       dh,
		Client:   client,
	}
}

func (app *App) NewMockRequest(request *http.Request) {
	app.Router.ServeHTTP(app.Recorder, request)
}
