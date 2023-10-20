package routes

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/authentication"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/handlers"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/models"
)

func GetLogin(c *gin.Context) {
	session := sessions.Default(c)
	flashes := session.Flashes()

	// clear flash messages
	session.Save()

	c.HTML(http.StatusOK, "login.html", gin.H{"flashes": flashes})
}

// TODO investigate HMAC error authentication -> seemingly not necessary due to gin HTML-escaping
func PostLogin(c *gin.Context, dh *handlers.DatabaseHandler) {
	session := sessions.Default(c)

	credentials := &models.Credentials{
		Username: c.PostForm("username"),
		Password: c.PostForm("password"),
	}

	id, e := authentication.ValidateCredentials(c, dh, credentials)
	if e != nil {
		log.Error().
			Err(e).
			Msg("Failed to validate credentials")

		session.AddFlash(e.Error())
		session.Save()

		// TODO make this a client side, no redirect operation
		c.Header("X-Redirect", "Forbidden")
		c.Redirect(http.StatusSeeOther, "login")
		return
	}

	log.Info().
		Str("user", *id).
		Msg("logged in")

	session.Set("user", credentials.Username)
	session.Set("id", id)
	session.Save()

	c.Header("X-Redirect", "Login")
	c.Redirect(http.StatusSeeOther, "admin/dashboard")
}
