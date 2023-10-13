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

	// this is not a precise mock of the behvior due to param injection
	// but the end-to-end behavior is exact
	request, e := http.NewRequest("GET", "/admin/dashboard/authenticated", nil)
	if e != nil {
		t.Fatal(e)
	}

	app.new_mock_request(request)

	// tests
	if status := app.recorder.Code; status != http.StatusOK {
		t.Errorf("Expected status code %v, but got %v", http.StatusOK, status)
	}
}

func Test_GetAdminDashboard_NoAuth_Fails(t *testing.T) {
	// initialize
	app := new_mock_app()
	defer app.database.Close(app.context)

	request, e := http.NewRequest("GET", "/admin/dashboard/notauthenticated", nil)
	if e != nil {
		t.Fatal(e)
	}

	app.new_mock_request(request)

	// tests
	if status := app.recorder.Code; status != http.StatusSeeOther {
		t.Errorf("Expected status code %v, but got %v", http.StatusOK, status)
	}

	header := app.recorder.Header()
	redirect := header.Get("X-Redirect")
	if redirect != "Forbidden" {
		t.Errorf("Expected header %s, but got %s", "Forbidden", redirect)
	}
}

func Test_GetChangePassword_Passes(t *testing.T) {
	// initialize
	app := new_mock_app()
	defer app.database.Close(app.context)

	request, e := http.NewRequest("GET", "/admin/password/authenticated", nil)
	if e != nil {
		t.Fatal(e)
	}

	app.new_mock_request(request)

	// tests
	if status := app.recorder.Code; status != http.StatusOK {
		t.Errorf("Expected status code %v, but got %v", http.StatusOK, status)
	}
}

func Test_GetChangePassword_NoAuth_Fails(t *testing.T) {
	// initialize
	app := new_mock_app()
	defer app.database.Close(app.context)

	request, e := http.NewRequest("GET", "/admin/password/notauthenticated", nil)
	if e != nil {
		t.Fatal(e)
	}

	app.new_mock_request(request)

	// tests
	if status := app.recorder.Code; status != http.StatusSeeOther {
		t.Errorf("Expected status code %v, but got %v", http.StatusOK, status)
	}

	header := app.recorder.Header()
	redirect := header.Get("X-Redirect")
	if redirect != "Forbidden" {
		t.Errorf("Expected header %s, but got %s", "Forbidden", redirect)
	}
}

func Test_PostChangePassword_Passes(t *testing.T) {
	// initialize
	app := new_mock_app()
	defer app.database.Close(app.context)

	prv_password := "user"
	new_password := "passwordthatislongerthan12characters"

	// Create a URL-encoded form data string
	data := url.Values{}
	data.Set("current_password", prv_password)
	data.Set("new_password", new_password)
	data.Set("new_password_confirm", new_password)
	form_data := data.Encode()

	// Create a POST request with the form data
	request, e := http.NewRequest("POST", "/admin/password/authenticated", strings.NewReader(form_data))
	if e != nil {
		t.Fatal(e)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	mock_id := uuid.NewString()
	prv_password_hash, _ := handlers.GeneratePHC(prv_password)
	app.database.ExpectQuery(`SELECT id, password_hash FROM users WHERE`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(
			pgxmock.NewRows([]string{"id", "password_hash"}).
				AddRow(mock_id, prv_password_hash),
		)

	app.database.ExpectExec(`UPDATE users SET`).
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	app.new_mock_request(request)

	// tests
	if status := app.recorder.Code; status != http.StatusSeeOther {
		t.Errorf("Expected status code %v, but got %v", http.StatusSeeOther, status)
	}
	header := app.recorder.Header()
	redirect := header.Get("X-Redirect")
	if redirect != "Password change" {
		t.Errorf("Expected header %s, but got %s", "Password change", redirect)
	}
}

func Test_PostChangePassword_UnconfirmedNewPassword_Fails(t *testing.T) {
	// initialize
	app := new_mock_app()
	defer app.database.Close(app.context)

	prv_password := "user"
	new_password := "passwordthatislongerthan12characters"

	// Create a URL-encoded form data string
	data := url.Values{}
	data.Set("current_password", prv_password)
	data.Set("new_password", new_password)
	data.Set("new_password_confirm", prv_password)
	form_data := data.Encode()

	// Create a POST request with the form data
	request, e := http.NewRequest("POST", "/admin/password/authenticated", strings.NewReader(form_data))
	if e != nil {
		t.Fatal(e)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	mock_id := uuid.NewString()
	prv_password_hash, _ := handlers.GeneratePHC(prv_password)
	app.database.ExpectQuery(`SELECT id, password_hash FROM users WHERE`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(
			pgxmock.NewRows([]string{"id", "password_hash"}).
				AddRow(mock_id, prv_password_hash),
		)

	app.new_mock_request(request)

	// tests
	if status := app.recorder.Code; status != http.StatusSeeOther {
		t.Errorf("Expected status code %v, but got %v", http.StatusSeeOther, status)
	}
	header := app.recorder.Header()
	redirect := header.Get("X-Redirect")
	if redirect != "Fields must match" {
		t.Errorf("Expected header %s, but got %s", "Fields must match", redirect)
	}
}

func Test_PostChangePassword_InvalidNewPassword_Fails(t *testing.T) {
	test_cases := []string{
		"tooshort",
		"toolong" + strings.Repeat("a", 128),
	}
	for _, tc := range test_cases {
		// initialize
		app := new_mock_app()
		defer app.database.Close(app.context)

		prv_password := "user"
		new_password := tc

		// Create a URL-encoded form data string
		data := url.Values{}
		data.Set("current_password", prv_password)
		data.Set("new_password", new_password)
		data.Set("new_password_confirm", new_password)
		form_data := data.Encode()

		// Create a POST request with the form data
		request, e := http.NewRequest("POST", "/admin/password/authenticated", strings.NewReader(form_data))
		if e != nil {
			t.Fatal(e)
		}
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		mock_id := uuid.NewString()
		prv_password_hash, _ := handlers.GeneratePHC(prv_password)
		app.database.ExpectQuery(`SELECT id, password_hash FROM users WHERE`).
			WithArgs(pgxmock.AnyArg()).
			WillReturnRows(
				pgxmock.NewRows([]string{"id", "password_hash"}).
					AddRow(mock_id, prv_password_hash),
			)

		app.new_mock_request(request)

		// tests
		if status := app.recorder.Code; status != http.StatusSeeOther {
			t.Errorf("Expected status code %v, but got %v", http.StatusSeeOther, status)
		}
		header := app.recorder.Header()
		redirect := header.Get("X-Redirect")
		if redirect != "Invalid password" {
			t.Errorf("Expected header %s, but got %s", "Invalid password", redirect)
		}
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
	admin.GET("/dashboard/:a", func(c *gin.Context) {
		a := c.Param("a")
		if a == "authenticated" {
			session := sessions.Default(c)
			session.Set("user", "user")
		}
		mock_admin_middleware(c)
		routes.GetAdminDashboard(c)
	})
	admin.GET("/password/:a", func(c *gin.Context) {
		a := c.Param("a")
		if a == "authenticated" {
			session := sessions.Default(c)
			session.Set("user", "user")
		}
		mock_admin_middleware(c)
		routes.GetChangePassword(c)
	})
	admin.POST("/password/:a", func(c *gin.Context) {
		a := c.Param("a")
		if a == "authenticated" {
			session := sessions.Default(c)
			session.Set("user", "user")
		}
		mock_admin_middleware(c)
		routes.PostChangePassword(c, rh)
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
