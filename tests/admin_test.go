package api_test

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v3"

	"github.com/solomonbaez/SB-Go-Newsletter-API/api/models"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/routes"
	adminRoutes "github.com/solomonbaez/SB-Go-Newsletter-API/api/routes/admin"
)

func TestAdmin(t *testing.T) {
	tests := []struct {
		name string
		fn   func(*testing.T)
	}{
		{"GetLogin", getLogin},
		{"PostLogin", postLogin},
		{"GetAdminDashboard", getAdminDashboard},
		{"GetChangePassword", getChangePassword},
		{"PostChangePassword", postChangePassword},
		{"GetLogout", getLogout},
	}

	for _, test := range tests {
		t.Run(test.name, test.fn)
	}
}

func getLogin(t *testing.T) {
	test := &struct {
		name           string
		expectedStatus int
	}{
		"(+) Test case -> GET request to /login -> passes",
		http.StatusOK,
	}

	t.Parallel()
	// initialize
	app := new_mock_app()
	app.router.GET("/login", routes.GetLogin)
	defer app.database.Close(app.context)

	request, e := http.NewRequest("GET", "/login", nil)
	if e != nil {
		t.Fatal(e)
	}

	app.new_mock_request(request)
	defer app.database.ExpectationsWereMet()

	// tests
	if returnedStatus := app.recorder.Code; returnedStatus != test.expectedStatus {
		t.Errorf("Expected status code %v, but got %v", test.expectedStatus, returnedStatus)
	}
}

// TODO research how to seed records into pgxmock tables
func postLogin(t *testing.T) {
	// base credentials to test against
	seedCredentials := &struct {
		userID       string
		username     string
		password     string
		passwordHash string
	}{
		userID:   uuid.NewString(),
		username: "user",
		password: "password",
	}
	seedCredentials.passwordHash, _ = models.GeneratePHC(seedCredentials.password)

	testCases := &[]struct {
		name           string
		username       string
		password       string
		expectedStatus int
		expectedHeader string
	}{
		{
			"(+) Test case 1 -> POST request to /login with valid credentials -> passes",
			"user",
			"password",
			http.StatusSeeOther,
			"Login",
		},
		{
			"(-) Test case 2 -> POST request to /login with invalid username -> fails",
			"resu",
			"password",
			http.StatusSeeOther,
			"Forbidden",
		},
		{
			"(-) Test case 3 -> POST request to /login with invalid password -> fails",
			"user",
			"drowssap",
			http.StatusSeeOther,
			"Forbidden",
		},
	}

	// parallelize tests
	t.Parallel()
	var app App
	for _, tc := range *testCases {
		// initialize
		app = new_mock_app()
		app.router.POST("/login", func(c *gin.Context) { routes.PostLogin(c, app.dh) })
		defer app.database.Close(app.context)

		// Create a URL-encoded form data string
		data := &url.Values{}
		data.Set("username", tc.username)
		data.Set("password", tc.password)
		form_data := data.Encode()

		request, e := http.NewRequest("POST", "/login", strings.NewReader(form_data))
		if e != nil {
			t.Fatal(e)
		}
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		query := app.database.ExpectQuery(`SELECT id, password_hash FROM users WHERE`).
			WithArgs(tc.username)
		if tc.username == seedCredentials.username {
			query.WillReturnRows(
				pgxmock.NewRows([]string{"id", "password_hash"}).
					AddRow(seedCredentials.userID, seedCredentials.passwordHash),
			)
		} else {
			query.WillReturnError(errors.New("Failed to validate credentials"))
		}

		app.new_mock_request(request)
		defer app.database.ExpectationsWereMet()

		// conditions
		if returnedStatus := app.recorder.Code; returnedStatus != tc.expectedStatus {
			t.Errorf("Expected status code %v, but got %v", tc.expectedStatus, returnedStatus)
		}

		returnedHeader := app.recorder.Header().Get("X-Redirect")
		if returnedHeader != tc.expectedHeader {
			t.Errorf("Expected header %s, but got %s", tc.expectedHeader, returnedHeader)
		}
	}
}

func getAdminDashboard(t *testing.T) {
	test := &struct {
		name           string
		expectedStatus int
	}{
		"(+) Test case -> GET request to /admin/dashboard -> passes",
		http.StatusOK,
	}

	t.Parallel()
	// initialize
	app := new_mock_app()
	admin = app.router.Group("/admin")
	admin.GET("/dashboard", adminRoutes.GetAdminDashboard)
	defer app.database.Close(app.context)

	// this is not a precise mock of the behvior due to param injection
	// but the end-to-end behavior is exact
	request, e := http.NewRequest("GET", "/admin/dashboard", nil)
	if e != nil {
		t.Fatal(e)
	}

	app.new_mock_request(request)
	defer app.database.ExpectationsWereMet()

	// tests
	if returnedStatus := app.recorder.Code; returnedStatus != test.expectedStatus {
		t.Errorf("Expected status code %v, but got %v", test.expectedStatus, returnedStatus)
	}
}

func getChangePassword(t *testing.T) {
	test := &struct {
		name           string
		expectedStatus int
	}{
		"(+) Test case -> GET request to /admin/dashboard -> passes",
		http.StatusOK,
	}

	t.Parallel()
	// initialize
	app := new_mock_app()
	admin = app.router.Group("/admin")
	admin.GET("/password", adminRoutes.GetChangePassword)
	defer app.database.Close(app.context)

	request, e := http.NewRequest("GET", "/admin/password", nil)
	if e != nil {
		t.Fatal(e)
	}

	app.new_mock_request(request)
	defer app.database.ExpectationsWereMet()

	// tests
	if returnedStatus := app.recorder.Code; returnedStatus != test.expectedStatus {
		t.Errorf("Expected status code %v, but got %v", test.expectedStatus, returnedStatus)
	}
}

func postChangePassword(t *testing.T) {
	// base credentials to test against
	seedCredentials := &struct {
		userID       string
		username     string
		password     string
		passwordHash string
	}{
		userID:   uuid.NewString(),
		username: "user",
		password: "password",
	}
	seedCredentials.passwordHash, _ = models.GeneratePHC(seedCredentials.password)

	testCases := &[]struct {
		name               string
		username           string
		password           string
		newPassword        string
		confirmNewPassword string
		expectedStatus     int
		expectedHeader     string
	}{
		{
			"(+) Test case 1 -> POST request to /admin/password with valid credentials and confirmed password-> passes",
			"user",
			"password",
			"passwordthatislongerthan12characters",
			"passwordthatislongerthan12characters",
			http.StatusSeeOther,
			"Password change",
		},
		{
			"(-) Test case 2 -> POST request to /admin/password with invalid username -> fails",
			"resu",
			"password",
			"passwordthatislongerthan12characters",
			"passwordthatislongerthan12characters",
			http.StatusSeeOther,
			"Forbidden",
		},
		{
			"(-) Test case 3 -> POST request to /admin/password with invalid password -> fails",
			"user",
			"drowssap",
			"passwordthatislongerthan12characters",
			"passwordthatislongerthan12characters",
			http.StatusSeeOther,
			"Forbidden",
		},
		{
			"(-) Test case 4 -> POST request to /admin/password with unconfirmed password -> fails",
			"user",
			"password",
			"unconfirmedpasswordthatislongerthan12characters",
			"passwordthatislongerthan12characters",
			http.StatusSeeOther,
			"Fields must match",
		},
		{
			"(-) Test case 5 -> POST request to /admin/password with password less than 12 characters -> fails",
			"user",
			"password",
			"tooshort",
			"passwordthatislongerthan12characters",
			http.StatusSeeOther,
			"Fields must match",
		},
		{
			"(-) Test case 5 -> POST request to /admin/password with password longer than 128 characters -> fails",
			"user",
			"password",
			"toolong" + strings.Repeat("a", 128),
			"passwordthatislongerthan12characters",
			http.StatusSeeOther,
			"Fields must match",
		},
	}

	// parallelize tests
	t.Parallel()
	var app App
	for _, tc := range *testCases {
		// initialize
		app = new_mock_app()
		defer app.database.Close(app.context)

		admin = app.router.Group("/admin")
		admin.POST("/password", func(c *gin.Context) { adminRoutes.PostChangePassword(c, app.dh) })

		// Create a URL-encoded form data string
		data := url.Values{}
		data.Set("current_password", tc.password)
		data.Set("new_password", tc.newPassword)
		data.Set("new_password_confirm", tc.confirmNewPassword)
		form_data := data.Encode()

		// Create a POST request with the form data
		request, e := http.NewRequest("POST", "/admin/password", strings.NewReader(form_data))
		if e != nil {
			t.Fatal(e)
		}
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		query := app.database.ExpectQuery(`SELECT id, password_hash FROM users WHERE`).
			WithArgs(pgxmock.AnyArg())
		if tc.username == seedCredentials.username {
			query.WillReturnRows(
				pgxmock.NewRows([]string{"id", "password_hash"}).
					AddRow(seedCredentials.userID, seedCredentials.passwordHash),
			)
		} else {
			query.WillReturnError(errors.New("Invalid credentials"))
		}

		app.database.ExpectExec(`UPDATE users SET`).
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		app.new_mock_request(request)
		defer app.database.ExpectationsWereMet()

		// tests
		if responseStatus := app.recorder.Code; responseStatus != tc.expectedStatus {
			t.Errorf("Expected status code %v, but got %v", tc.expectedStatus, responseStatus)
		}
		responseHeader := app.recorder.Header().Get("X-Redirect")
		if responseHeader != tc.expectedHeader {
			t.Errorf("Expected header %s, but got %s", tc.expectedHeader, responseHeader)
		}
	}
}

func getLogout(t *testing.T) {
	test := &struct {
		name           string
		expectedStatus int
		expectedHeader string
	}{
		"(+) Test case -> GET request to /admin/logout -> passes",
		http.StatusSeeOther,
		"Logged out",
	}

	t.Parallel()
	// initialize
	app := new_mock_app()
	admin = app.router.Group("/admin")
	admin.GET("/logout", adminRoutes.Logout)
	defer app.database.Close(app.context)

	request, e := http.NewRequest("GET", "/admin/logout", nil)
	if e != nil {
		t.Fatal(e)
	}

	app.new_mock_request(request)
	defer app.database.ExpectationsWereMet()

	// tests
	if responseStatus := app.recorder.Code; responseStatus != test.expectedStatus {
		t.Errorf("Expected status code %v, but got %v", test.expectedStatus, responseStatus)
	}

	responseHeader := app.recorder.Header().Get("X-Redirect")
	if responseHeader != test.expectedHeader {
		t.Errorf("Expected header %s, but got %s", test.expectedHeader, responseHeader)
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
