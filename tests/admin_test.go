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
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/handlers"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/models"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/routes"
	adminRoutes "github.com/solomonbaez/SB-Go-Newsletter-API/api/routes/admin"
)

// Reference
type App struct {
	recorder *httptest.ResponseRecorder
	context  *gin.Context
	router   *gin.Engine
	database pgxmock.PgxConnIface
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
	mock_password_hash, _ := models.GeneratePHC(mock_password)
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
	invalid_password_hash, _ := models.GeneratePHC(invalid_password)
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
		t.Errorf("Expected status code %v, but got %v", http.StatusSeeOther, status)
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
		t.Errorf("Expected status code %v, but got %v", http.StatusSeeOther, status)
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
	prv_password_hash, _ := models.GeneratePHC(prv_password)
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
	data.Set("new_password_confirm", "")
	form_data := data.Encode()

	// Create a POST request with the form data
	request, e := http.NewRequest("POST", "/admin/password/authenticated", strings.NewReader(form_data))
	if e != nil {
		t.Fatal(e)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	mock_id := uuid.NewString()
	prv_password_hash, _ := models.GeneratePHC(prv_password)
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
		prv_password_hash, _ := models.GeneratePHC(prv_password)
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

func Test_GetLogout_Passes(t *testing.T) {
	// initialize
	app := new_mock_app()
	defer app.database.Close(app.context)

	request, e := http.NewRequest("GET", "/admin/logout/authenticated", nil)
	if e != nil {
		t.Fatal(e)
	}

	app.new_mock_request(request)

	// tests
	if status := app.recorder.Code; status != http.StatusSeeOther {
		t.Errorf("Expected status code %v, but got %v", http.StatusSeeOther, status)
	}

	header := app.recorder.Header()
	redirect := header.Get("X-Redirect")
	if redirect != "Logged out" {
		t.Errorf("Expected header %s, but got %s", "Logged out", redirect)
	}
}

func new_mock_database() (database pgxmock.PgxConnIface) {
	database, _ = pgxmock.NewConn()

	return database
}

// func new_mock_client() (client *clients.SMTPClient) {
// 	cfg := "test"
// 	client, _ = clients.NewSMTPClient(&cfg)

// 	return client
// }

func new_mock_app() App {
	var recorder *httptest.ResponseRecorder
	var context *gin.Context
	var database pgxmock.PgxConnIface
	var dh *handlers.DatabaseHandler
	var store cookie.Store
	var admin *gin.RouterGroup

	recorder = httptest.NewRecorder()
	database = new_mock_database()
	dh = handlers.NewDatabaseHandler(database)

	router := gin.Default()
	context = gin.CreateTestContextOnly(recorder, router)
	router.LoadHTMLGlob(filepath.Join("../api/templates", "*"))

	store = cookie.NewStore([]byte("test"))
	router.Use(sessions.Sessions("test", store))

	router.GET("/login", routes.GetLogin)
	router.POST("/login", func(c *gin.Context) {
		routes.PostLogin(c, dh)
	})

	admin = router.Group("/admin")
	admin.GET("/dashboard/:a", func(c *gin.Context) {
		mock_login(c)
		mock_admin_middleware(c)
		adminRoutes.GetAdminDashboard(c)
	})
	admin.GET("/password/:a", func(c *gin.Context) {
		mock_login(c)
		mock_admin_middleware(c)
		adminRoutes.GetChangePassword(c)
	})
	admin.POST("/password/:a", func(c *gin.Context) {
		mock_login(c)
		mock_admin_middleware(c)
		adminRoutes.PostChangePassword(c, dh)
	})
	admin.GET("/logout/:a", func(c *gin.Context) {
		mock_login(c)
		mock_admin_middleware(c)
		adminRoutes.Logout(c)
	})

	return App{
		recorder: recorder,
		context:  context,
		router:   router,
		database: database,
	}
}

func (app *App) new_mock_request(request *http.Request) {
	app.router.ServeHTTP(app.recorder, request)
}

func mock_login(c *gin.Context) {
	a := c.Param("a")
	if a == "authenticated" {
		session := sessions.Default(c)
		session.Set("user", "user")
	}
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

// // TODO NEED TO BE AMENDED
// func Test_PostNewsletter_Passes(t *testing.T) {
// 	// initialize
// 	database, e := spawn_mock_database()
// 	if e != nil {
// 		t.Fatal(e)
// 	}

// 	cfg := mock.ConfigurationAttr{}
// 	server := mock.New(cfg)
// 	server.Start()
// 	defer server.Stop()
// 	port := server.PortNumber

// 	client, e := spawn_mock_smtp_client()
// 	if e != nil {
// 		t.Fatal(e)
// 	}
// 	client.SmtpServer = "[::]"
// 	client.SmtpPort = port
// 	client.Sender = models.SubscriberEmail("user@test.com")

// 	body := models.Body{
// 		Title: "testing",
// 		Text:  "testing",
// 		Html:  "<p>testing</p>",
// 	}
// 	emailContent := models.Newsletter{
// 		Recipient: models.SubscriberEmail("recipient@test.com"),
// 		Content:   &body,
// 	}
// 	fmt.Printf(emailContent.Content.Html)

// 	router := spawn_mock_router(database, client)

// 	mock_username := "user"
// 	mock_password := "password"
// 	data := `{"title":"test", "text":"test", "html":"<p>test</p>"}`
// 	request, e := http.NewRequest("POST", "/newsletter", strings.NewReader(data))
// 	if e != nil {
// 		t.Fatal(e)
// 	}
// 	request.SetBasicAuth(mock_username, mock_password)

// 	mock_id := uuid.NewString()
// 	mock_password_hash, e := handlers.GeneratePHC(mock_password)
// 	if e != nil {
// 		t.Fatal(e)
// 	}
// 	database.ExpectQuery(`SELECT id, password_hash FROM users WHERE`).
// 		WithArgs(pgxmock.AnyArg()).
// 		WillReturnRows(
// 			pgxmock.NewRows([]string{"id", "password_hash"}).
// 				AddRow(mock_id, mock_password_hash),
// 		)

// 	database.ExpectQuery(`SELECT id, email, name, created, status FROM subscriptions WHERE`).
// 		WithArgs(pgxmock.AnyArg()).
// 		WillReturnRows(
// 			pgxmock.NewRows([]string{"id", "email", "name", "created", "status"}).
// 				AddRow(mock_id, models.SubscriberEmail("test@test.com"), models.SubscriberName("TestUser"), time.Now(), "confirmed"),
// 		)

// 	app := spawn_app(router, request)
// 	defer database.ExpectationsWereMet()
// 	defer database.Close(app.context)

// 	// tests
// 	if status := app.recorder.Code; status != http.StatusOK {
// 		t.Errorf("Expected status code %v, but got %v", http.StatusOK, status)
// 	}

// 	expected_body := fmt.Sprintf(
// 		`{"requestID":"","subscribers":[{"id":"%s","email":"test@test.com","name":"TestUser","status":"confirmed"}]}`, mock_id,
// 	) + `{"message":"Emails successfully delivered","requestID":""}`
// 	response_body := app.recorder.Body.String()
// 	if response_body != expected_body {
// 		t.Errorf("Expected body %v, but got %v", expected_body, response_body)
// 	}
// }

// func Test_PostNewsletter_InvalidPassword_Fails(t *testing.T) {
// 	// initialize
// 	database, e := spawn_mock_database()
// 	if e != nil {
// 		t.Fatal(e)
// 	}

// 	cfg := mock.ConfigurationAttr{}
// 	server := mock.New(cfg)
// 	server.Start()
// 	defer server.Stop()
// 	port := server.PortNumber

// 	client, e := spawn_mock_smtp_client()
// 	if e != nil {
// 		t.Fatal(e)
// 	}
// 	client.SmtpServer = "[::]"
// 	client.SmtpPort = port
// 	client.Sender = models.SubscriberEmail("user@test.com")

// 	body := models.Body{
// 		Title: "testing",
// 		Text:  "testing",
// 		Html:  "<p>testing</p>",
// 	}
// 	emailContent := models.Newsletter{
// 		Recipient: models.SubscriberEmail("recipient@test.com"),
// 		Content:   &body,
// 	}
// 	fmt.Printf(emailContent.Content.Html)

// 	router := spawn_mock_router(database, client)

// 	mock_username := "user"
// 	mock_password := "password"
// 	invalid_password := "drowssap"
// 	data := `{"title":"test", "text":"test", "html":"<p>test</p>"}`
// 	request, e := http.NewRequest("POST", "/newsletter", strings.NewReader(data))
// 	if e != nil {
// 		t.Fatal(e)
// 	}
// 	// I'm relatively unconcerned about basic auth failing in this integration test
// 	// TODO sketch out a unit test for handlers.BasicAuth
// 	request.SetBasicAuth(mock_username, mock_password)

// 	mock_id := uuid.NewString()
// 	invalid_password_hash, e := handlers.GeneratePHC(invalid_password)
// 	if e != nil {
// 		t.Fatal(e)
// 	}
// 	database.ExpectQuery(`SELECT id, password_hash FROM users WHERE`).
// 		WithArgs(pgxmock.AnyArg()).
// 		WillReturnRows(
// 			pgxmock.NewRows([]string{"id", "password_hash"}).
// 				AddRow(mock_id, invalid_password_hash),
// 		)

// 	app := spawn_app(router, request)
// 	defer database.ExpectationsWereMet()
// 	defer database.Close(app.context)

// 	// tests
// 	if status := app.recorder.Code; status != http.StatusBadRequest {
// 		t.Errorf("Expected status code %v, but got %v", http.StatusOK, status)
// 	}

// 	expected_body := `{"error":"Failed to validate credentials: PHC are not equivalent","requestID":""}`
// 	response_body := app.recorder.Body.String()
// 	if response_body != expected_body {
// 		t.Errorf("Expected body %v, but got %v", expected_body, response_body)
// 	}
// }

// func Test_PostNewsletter_InvalidNewsletter_Fails(t *testing.T) {
// 	// // initialization
// 	var database pgxmock.PgxConnIface
// 	var client *clients.SMTPClient
// 	var router *gin.Engine
// 	var request *http.Request
// 	var app app
// 	var e error

// 	cfg := mock.ConfigurationAttr{}
// 	server := mock.New(cfg)

// 	mock_username := "user"
// 	mock_password := "password"
// 	mock_id := uuid.NewString()

// 	test_cases := []*models.Body{
// 		{
// 			Title: "",
// 			Text:  "testing",
// 			Html:  "<p>testing</p>",
// 		},
// 		{
// 			Title: "testing",
// 			Text:  "",
// 			Html:  "<p>testing</p>",
// 		},
// 		{
// 			Title: "testing",
// 			Text:  "testing",
// 			Html:  "",
// 		},
// 	}
// 	expected_responses := []string{
// 		`{"error":"Failed to parse newsletter: field: Title cannot be empty","requestID":""}`,
// 		`{"error":"Failed to parse newsletter: field: Text cannot be empty","requestID":""}`,
// 		`{"error":"Failed to parse newsletter: field: Html cannot be empty","requestID":""}`,
// 	}

// 	for i, tc := range test_cases {
// 		// initialize
// 		database, e = spawn_mock_database()
// 		if e != nil {
// 			t.Fatal(e)
// 		}

// 		server.Start()
// 		defer server.Stop()
// 		port := server.PortNumber

// 		client, e = spawn_mock_smtp_client()
// 		if e != nil {
// 			t.Fatal(e)
// 		}
// 		client.SmtpServer = "[::]"
// 		client.SmtpPort = port
// 		client.Sender = models.SubscriberEmail("user@test.com")

// 		router = spawn_mock_router(database, client)

// 		mock_password_hash, e := handlers.GeneratePHC(mock_password)
// 		if e != nil {
// 			t.Fatal(e)
// 		}

// 		data := fmt.Sprintf(`{"title":"%s", "text":"%s", "html":"%s"}`, tc.Title, tc.Text, tc.Html)
// 		request, e = http.NewRequest("POST", "/newsletter", strings.NewReader(data))
// 		if e != nil {
// 			t.Fatal(e)
// 		}
// 		request.SetBasicAuth(mock_username, mock_password)

// 		database.ExpectQuery(`SELECT id, password_hash FROM users WHERE`).
// 			WithArgs(pgxmock.AnyArg()).
// 			WillReturnRows(
// 				pgxmock.NewRows([]string{"id", "password_hash"}).
// 					AddRow(mock_id, mock_password_hash),
// 			)

// 		database.ExpectQuery(`SELECT id, email, name, created, status FROM subscriptions WHERE`).
// 			WithArgs(pgxmock.AnyArg()).
// 			WillReturnRows(
// 				pgxmock.NewRows([]string{"id", "email", "name", "created", "status"}).
// 					AddRow(mock_id, models.SubscriberEmail("test@test.com"), models.SubscriberName("TestUser"), time.Now(), "confirmed"),
// 			)

// 		app = spawn_app(router, request)
// 		defer database.ExpectationsWereMet()
// 		defer database.Close(app.context)

// 		// tests
// 		if status := app.recorder.Code; status != http.StatusBadRequest {
// 			t.Errorf("Expected status code %v, but got %v", http.StatusOK, status)
// 		}

// 		response_body := app.recorder.Body.String()
// 		if response_body != expected_responses[i] {
// 			t.Errorf("Expected body %v, but got %v", expected_responses[i], response_body)
// 		}
// 	}
// }
