package routes

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func GetAdminDashboard(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get("user")
	session.Save()

	c.HTML(http.StatusOK, "dashboard.html", gin.H{"user": user})
}
