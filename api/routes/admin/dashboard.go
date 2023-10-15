package routes

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func GetAdminDashboard(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get("user")
	flashes := session.Flashes()
	session.Save()

	c.HTML(http.StatusOK, "dashboard.html", gin.H{"flashes": flashes, "user": user})
}
