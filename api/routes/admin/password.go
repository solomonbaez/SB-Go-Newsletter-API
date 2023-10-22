package routes

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/authentication"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/handlers"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/models"
)

func GetChangePassword(c *gin.Context) {
	session := sessions.Default(c)
	flashes := session.Flashes()
	session.Save()

	c.HTML(http.StatusOK, "password.html", gin.H{"flashes": flashes})
}

func PostChangePassword(c *gin.Context, dh *handlers.DatabaseHandler) {
	session := sessions.Default(c)

	u := session.Get("user")
	user := fmt.Sprintf("%v", u)
	credentials := &models.Credentials{
		Username: user,
		Password: c.PostForm("current_password"),
	}

	id, e := authentication.ValidateCredentials(c, dh, credentials)
	if e != nil {
		session.AddFlash(e.Error())
		session.Save()

		c.Header("X-Redirect", "Forbidden")
		c.Redirect(http.StatusSeeOther, "password")
		return
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

		c.Header("X-Redirect", "Fields must match")
		c.Redirect(http.StatusSeeOther, "password")
		return
	}
	if e = ParsePassword(newPassword); e != nil {
		log.Error().
			Err(e).
			Msg("Failed to parse password")

		session.AddFlash(e.Error())
		session.Save()

		c.Header("X-Redirect", "Invalid password")
		c.Redirect(http.StatusSeeOther, "password")
		return
	}

	newPHC, e := models.GeneratePHC(newPassword)
	if e != nil {
		log.Error().
			Err(e).
			Msg("Failed to generate new PHC")

		session.AddFlash(e.Error())
		session.Save()

		c.Header("X-Redirect", e.Error())

		c.Redirect(http.StatusSeeOther, "password")
		return
	}

	if e = ChangePassword(c, dh, id, newPHC); e != nil {
		log.Error().
			Err(e).
			Msg("Failed to change password")

		session.AddFlash(e.Error())
		session.Save()

		c.Header("X-Redirect", e.Error())

		c.Redirect(http.StatusSeeOther, "password")
		return
	}

	log.Info().
		Str("user", user).
		Msg("Password has been changed")

	c.Header("X-Redirect", "Password change")

	session.AddFlash("Password has been changed")
	session.Save()
	c.Redirect(http.StatusSeeOther, "dashboard")
}

func ChangePassword(c *gin.Context, dh *handlers.DatabaseHandler, id *string, newPHC string) (err error) {
	query := "UPDATE users SET password_hash = $1 WHERE id = $2"
	_, e := dh.DB.Exec(c, query, newPHC, id)
	if e != nil {
		err = fmt.Errorf("failed change password: %w", e)
		return
	}

	return
}

// ParseField dependency injection
func ParsePassword(password string) (err error) {
	// sanitize password
	if _, e := authentication.ParseField(password); e != nil {
		err = fmt.Errorf("failed to sanitize password: %w", e)
		return
	}

	// parse password length per OWASP minimum requirements
	if len(password) <= 12 {
		err = errors.New("password must be longer than 12 characters")
		return
	}
	if len(password) >= 128 {
		err = errors.New("password must be shorter than 128 characters")
		return
	}

	return
}
