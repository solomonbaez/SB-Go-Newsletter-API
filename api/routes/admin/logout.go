package routes

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.AddFlash("logged out")
	session.Save()

	c.Header("X-Redirect", "Logged out")

	c.Redirect(http.StatusSeeOther, "../login")
}
