package api_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"

	"testing"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

	request, e := http.NewRequest("GET", "/login", nil)
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

	mock_username := "user"
	mock_password := "password"

	// Create a URL-encoded form data string
	data := url.Values{}
	data.Set("username", mock_username)
	data.Set("password", mock_password)
	form_data := data.Encode()

	// Create a POST request with the form data
	request, e := http.NewRequest("POST", "/login", strings.NewReader(form_data))
	if e != nil {
		t.Fatal(e)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	mock_id := uuid.NewString()
	mock_password_hash, _ := handlers.GeneratePHC(mock_password)
	app.database.ExpectQuery(`SELECT id, password_hash FROM users WHERE`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(
			pgxmock.NewRows([]string{"id", "password_hash"}).
				AddRow(mock_id, mock_password_hash),
		)

	app.new_mock_request(request)

	// tests
	if status := app.recorder.Code; status != http.StatusSeeOther {
		t.Errorf("Expected status code %v, but got %v", http.StatusSeeOther, status)
	}

	header := app.recorder.Header()
	redirect := header.Get("X-Redirect")
	if redirect != "Login" {
		t.Errorf("Expected header %s, but got %s", "Login", redirect)
	}
}

func Test_PostLogin_InvalidCredentials_Fails(t *testing.T) {
	// initialize
	app := new_mock_app()
	defer app.database.Close(app.context)

	mock_username := "user"
	mock_password := "password"
	invalid_password := "drowssap"

	// Create a URL-encoded form data string
	data := url.Values{}
	data.Set("username", mock_username)
	data.Set("password", mock_password)
	form_data := data.Encode()

	// Create a POST request with the form data
	request, e := http.NewRequest("POST", "/login", strings.NewReader(form_data))
	if e != nil {
		t.Fatal(e)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	mock_id := uuid.NewString()
	invalid_password_hash, _ := handlers.GeneratePHC(invalid_password)
	app.database.ExpectQuery(`SELECT id, password_hash FROM users WHERE`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(
			pgxmock.NewRows([]string{"id", "password_hash"}).
				AddRow(mock_id, invalid_password_hash),
		)

	app.new_mock_request(request)

	// tests
	if status := app.recorder.Code; status != http.StatusSeeOther {
		t.Errorf("Expected status code %v, but got %v", http.StatusSeeOther, status)
	}

	header := app.recorder.Header()
	redirect := header.Get("X-Redirect")
	if redirect != "Forbidden" {
		t.Errorf("Expected header %s, but got %s", "Forbidden", redirect)
	}
}

func Test_GetAdminDashboard_Passes(t *testing.T) {
	// initialize
	app := new_mock_app()
	defer app.database.Close(app.context)

	request, e := http.NewRequest("GET", "/admin/dashboard", nil)
	if e != nil {
		t.Fatal(e)
	}

	app.new_mock_request(request)

	// tests
	if status := app.recorder.Code; status != http.StatusOK {
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
	router.GET("/login", routes.GetLogin)
	router.POST("/login", func(c *gin.Context) {
		routes.PostLogin(c, rh)
	})

	admin := router.Group("/admin")
	admin.GET("/dashboard", func(c *gin.Context) {
		// mock middleware behavior
		session := sessions.Default(c)
		session.Set("user", "user")
		mock_admin_middleware(c)
		routes.GetAdminDashboard(c)
	})

	recorder := httptest.NewRecorder()
	app.recorder = recorder

	context := gin.CreateTestContextOnly(recorder, router)
	app.context = context

	app = App{
		recorder: recorder,
		context:  context,
		router:   router,
		database: database,
		client:   client,
	}

	return app
}

func (app *App) new_mock_request(request *http.Request) {
	app.router.ServeHTTP(app.recorder, request)
}

func mock_admin_middleware(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get("user")
	if user == nil {
		c.Header("X-Redirect", "Forbidden")
		c.Redirect(http.StatusSeeOther, "../login")
		c.Abort()
		return
	}
}
