package api_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/clients"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/handlers"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/idempotency"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/routes"
	adminRoutes "github.com/solomonbaez/SB-Go-Newsletter-API/api/routes/admin"
)

type App struct {
	recorder *httptest.ResponseRecorder
	context  *gin.Context
	router   *gin.Engine
	database pgxmock.PgxConnIface
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

func new_mock_app() App {
	var recorder *httptest.ResponseRecorder
	var context *gin.Context
	var database pgxmock.PgxConnIface
	var client *clients.SMTPClient
	var dh *handlers.DatabaseHandler
	var store cookie.Store
	var admin *gin.RouterGroup

	recorder = httptest.NewRecorder()
	database = new_mock_database()
	client = new_mock_client()
	dh = handlers.NewDatabaseHandler(database)

	router := gin.Default()
	context = gin.CreateTestContextOnly(recorder, router)
	router.LoadHTMLGlob(filepath.Join("../api/templates", "*"))

	store = cookie.NewStore([]byte("test"))
	router.Use(sessions.Sessions("test", store))

	router.GET("/health", handlers.HealthCheck)
	router.GET("/home", routes.Home)
	router.GET("/login", routes.GetLogin)
	router.POST("/login", func(c *gin.Context) { routes.PostLogin(c, dh) })
	router.POST("/subscribe", func(c *gin.Context) { routes.Subscribe(c, dh, client) })
	router.GET("/confirm/:token", func(c *gin.Context) { routes.ConfirmSubscriber(c, dh) })

	admin = router.Group("/admin")
	admin.GET("/dashboard", adminRoutes.GetAdminDashboard)
	admin.GET("/password", adminRoutes.GetChangePassword)
	admin.POST("/password", func(c *gin.Context) { adminRoutes.PostChangePassword(c, dh) })
	admin.GET("/logout", adminRoutes.Logout)
	admin.GET("/subscribers", func(c *gin.Context) { adminRoutes.GetSubscribers(c, dh) })
	admin.GET("/subscribers/:id", func(c *gin.Context) { adminRoutes.GetSubscriberByID(c, dh) })
	admin.GET("/confirmed", func(c *gin.Context) { _ = adminRoutes.GetConfirmedSubscribers(c, dh) })
	admin.GET("/newsletter", adminRoutes.GetNewsletter)
	admin.POST("/newsletter", func(c *gin.Context) { adminRoutes.PostNewsletter(c, dh, client) })
	admin.GET("/responses", func(c *gin.Context) {
		session := sessions.Default(c)
		id := fmt.Sprintf("%s", session.Get("user"))
		key := fmt.Sprintf("%s", session.Get("key"))
		idempotency.GetSavedResponse(c, dh, id, key)
	})

	admin.POST("/responses", func(c *gin.Context) {
		response, _ := adminRoutes.SeeOther(c, "admin/dashboard")
		idempotency.SaveResponse(c, dh, response)
	})

	return App{
		recorder: recorder,
		context:  context,
		router:   router,
		database: database,
		client:   client,
	}
}

func (app *App) new_mock_request(request *http.Request) {
	app.router.ServeHTTP(app.recorder, request)
}
