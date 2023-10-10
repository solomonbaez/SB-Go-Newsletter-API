package routes

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func GetAdminDashboard(c *gin.Context) {
	s := sessions.Default(c)
	session := RotateSession(c, s)
	user := session.Get("user")

	c.HTML(http.StatusOK, "dashboard.html", gin.H{"user": user})
}

func RotateSession(c *gin.Context, prv sessions.Session) sessions.Session {
	user := prv.Get("user")
	if user == nil {
		c.Redirect(http.StatusSeeOther, "../login")
	}

	new := sessions.Default(c)
	new.Set("user", user)
	new.Save()
	prv.Clear()

	return new
}
