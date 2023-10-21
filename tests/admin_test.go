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
	utils "github.com/solomonbaez/SB-Go-Newsletter-API/test_utils"
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

	t.Parallel()
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
	app := utils.NewMockApp()
	app.Router.GET("/login", routes.GetLogin)
	defer app.Database.Close(app.Context)

	request, _ := http.NewRequest("GET", "/login", nil)

	app.NewMockRequest(request)
	defer app.Database.ExpectationsWereMet()

	// tests
	if returnedStatus := app.Recorder.Code; returnedStatus != test.expectedStatus {
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
	var app utils.App
	for _, tc := range *testCases {
		// initialize
		app = utils.NewMockApp()
		app.Router.POST("/login", func(c *gin.Context) { routes.PostLogin(c, app.DH) })
		defer app.Database.Close(app.Context)

		// Create a URL-encoded form data string
		data := &url.Values{}
		data.Set("username", tc.username)
		data.Set("password", tc.password)
		form_data := data.Encode()

		request, _ := http.NewRequest("POST", "/login", strings.NewReader(form_data))
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		query := app.Database.ExpectQuery(`SELECT id, password_hash FROM users WHERE`).
			WithArgs(tc.username)
		if tc.username == seedCredentials.username {
			query.WillReturnRows(
				pgxmock.NewRows([]string{"id", "password_hash"}).
					AddRow(seedCredentials.userID, seedCredentials.passwordHash),
			)
		} else {
			query.WillReturnError(errors.New("Failed to validate credentials"))
		}

		app.NewMockRequest(request)
		defer app.Database.ExpectationsWereMet()

		// conditions
		if returnedStatus := app.Recorder.Code; returnedStatus != tc.expectedStatus {
			t.Errorf("Expected status code %v, but got %v", tc.expectedStatus, returnedStatus)
		}

		returnedHeader := app.Recorder.Header().Get("X-Redirect")
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
	app := utils.NewMockApp()
	admin := app.Router.Group("/admin")
	admin.GET("/dashboard", adminRoutes.GetAdminDashboard)
	defer app.Database.Close(app.Context)

	// this is not a precise mock of the behvior due to param injection
	// but the end-to-end behavior is exact
	request, _ := http.NewRequest("GET", "/admin/dashboard", nil)

	app.NewMockRequest(request)
	defer app.Database.ExpectationsWereMet()

	// tests
	if returnedStatus := app.Recorder.Code; returnedStatus != test.expectedStatus {
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
	app := utils.NewMockApp()
	admin := app.Router.Group("/admin")
	admin.GET("/password", adminRoutes.GetChangePassword)
	defer app.Database.Close(app.Context)

	request, _ := http.NewRequest("GET", "/admin/password", nil)

	app.NewMockRequest(request)
	defer app.Database.ExpectationsWereMet()

	// tests
	if returnedStatus := app.Recorder.Code; returnedStatus != test.expectedStatus {
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
	var app utils.App
	for _, tc := range *testCases {
		// initialize
		app = utils.NewMockApp()
		defer app.Database.Close(app.Context)

		admin := app.Router.Group("/admin")
		admin.POST("/password", func(c *gin.Context) { adminRoutes.PostChangePassword(c, app.DH) })

		// Create a URL-encoded form data string
		data := url.Values{}
		data.Set("current_password", tc.password)
		data.Set("new_password", tc.newPassword)
		data.Set("new_password_confirm", tc.confirmNewPassword)
		form_data := data.Encode()

		// Create a POST request with the form data
		request, _ := http.NewRequest("POST", "/admin/password", strings.NewReader(form_data))

		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		query := app.Database.ExpectQuery(`SELECT id, password_hash FROM users WHERE`).
			WithArgs(pgxmock.AnyArg())
		if tc.username == seedCredentials.username {
			query.WillReturnRows(
				pgxmock.NewRows([]string{"id", "password_hash"}).
					AddRow(seedCredentials.userID, seedCredentials.passwordHash),
			)
		} else {
			query.WillReturnError(errors.New("Invalid credentials"))
		}

		app.Database.ExpectExec(`UPDATE users SET`).
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		app.NewMockRequest(request)
		defer app.Database.ExpectationsWereMet()

		// tests
		if responseStatus := app.Recorder.Code; responseStatus != tc.expectedStatus {
			t.Errorf("Expected status code %v, but got %v", tc.expectedStatus, responseStatus)
		}
		responseHeader := app.Recorder.Header().Get("X-Redirect")
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
	app := utils.NewMockApp()
	admin := app.Router.Group("/admin")
	admin.GET("/logout", adminRoutes.Logout)
	defer app.Database.Close(app.Context)

	request, _ := http.NewRequest("GET", "/admin/logout", nil)

	app.NewMockRequest(request)
	defer app.Database.ExpectationsWereMet()

	// tests
	if responseStatus := app.Recorder.Code; responseStatus != test.expectedStatus {
		t.Errorf("Expected status code %v, but got %v", test.expectedStatus, responseStatus)
	}

	responseHeader := app.Recorder.Header().Get("X-Redirect")
	if responseHeader != test.expectedHeader {
		t.Errorf("Expected header %s, but got %s", test.expectedHeader, responseHeader)
	}
}
