package routes

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/handlers"
)

func GetLogin(c *gin.Context) {
	session := sessions.Default(c)
	flashes := session.Flashes()

	// clear flash messages
	session.Save()

	c.HTML(http.StatusOK, "login.html", gin.H{"flashes": flashes})
}

// TODO investigate HMAC error authentication -> seemingly not necessary due to gin HTML-escaping
func PostLogin(c *gin.Context, rh *handlers.RouteHandler) {
	session := sessions.Default(c)

	credentials := &handlers.Credentials{
		Username: c.PostForm("username"),
		Password: c.PostForm("password"),
	}

	id, e := rh.ValidateCredentials(c, credentials)
	if e != nil {
		log.Error().
			Err(e).
			Msg("Failed to validate credentials")

		session.AddFlash(e.Error())
		session.Save()

		c.Redirect(http.StatusSeeOther, "login")
	} else {
		log.Info().
			Str("id", *id).
			Msg("login")

		session.Set("user", credentials.Username)

		c.Redirect(http.StatusSeeOther, "admin/dashboard")
	}
}
