package routes

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/handlers"
)

func GetChangePassword(c *gin.Context) {
	session := sessions.Default(c)
	flashes := session.Flashes()

	c.HTML(http.StatusOK, "password.html", gin.H{"flashes": flashes})
}

func PostChangePassword(c *gin.Context, rh *handlers.RouteHandler) {
	session := sessions.Default(c)

	u := session.Get("user")
	user := fmt.Sprintf("%v", u)

	credentials := &handlers.Credentials{
		Username: user,
		Password: c.PostForm("current_password"),
	}
	id, e := rh.ValidateCredentials(c, credentials)
	if e != nil {
		session.AddFlash(e.Error())
		session.Save()

		c.Redirect(http.StatusSeeOther, "password")
	}

	newPassword := c.PostForm("new_password")
	confirm := c.PostForm("new_password_confirm")
	if newPassword != confirm {
		e := errors.New("new password fields must match")
		log.Error().
			Err(e).
			Msg("Failed to validate new password")

		session.AddFlash(e.Error())
		session.Save()

		c.Redirect(http.StatusSeeOther, "password")
	}
	newPHC, e := handlers.GeneratePHC(newPassword)
	if e != nil {
		log.Error().
			Err(e).
			Msg("Failed to generate new PHC")

		session.AddFlash(e.Error())
		session.Save()

		c.Redirect(http.StatusSeeOther, "password")
	}

	if e = ChangePassword(c, rh, id, newPHC); e != nil {
		log.Error().
			Err(e).
			Msg("Failed to change password")

		session.AddFlash(e.Error())
		session.Save()

		c.Redirect(http.StatusSeeOther, "password")
	}

	log.Info().
		Str("user", user).
		Msg("Password has been changed")

	session.AddFlash("Password has been changed")
	session.Save()
	c.Redirect(http.StatusSeeOther, "dashboard")
}

func ChangePassword(c *gin.Context, rh *handlers.RouteHandler, id *string, newPHC string) (e error) {
	query := "UPDATE users SET password_hash = $1 WHERE id = $2"
	_, e = rh.DB.Exec(c, query, newPHC, id)
	if e != nil {
		return e
	}

	return
}
