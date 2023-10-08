package routes

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/handlers"
)

func GetLogin(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{"title": "login"})
}

// TODO investigate HMAC error authentication -> seemingly not necessary due to gin HTML-escaping
func PostLogin(c *gin.Context, rh *handlers.RouteHandler) {
	credentials := &handlers.Credentials{
		Username: c.PostForm("username"),
		Password: c.PostForm("password"),
	}

	id, e := rh.ValidateCredentials(c, credentials)
	if e != nil {
		log.Error().
			Err(e).
			Msg("Failed to validate credentials")

		c.HTML(http.StatusSeeOther, "login.html", gin.H{"error": fmt.Sprintf("Error: %s", e.Error())})
	} else {
		log.Info().
			Str("id", *id).
			Msg("login")

		c.Redirect(http.StatusSeeOther, "home")
	}
}
