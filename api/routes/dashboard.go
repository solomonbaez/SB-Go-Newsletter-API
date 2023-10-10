package routes

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	// "github.com/rs/zerolog/log"
	// "github.com/solomonbaez/SB-Go-Newsletter-API/api/handlers"
)

func GetAdminDashboard(c *gin.Context) {
	session := sessions.Default(c)
	id := session.Get("user")
	if id == nil {
		c.Redirect(http.StatusSeeOther, "../login")
	}

	c.HTML(http.StatusOK, "dashboard.html", gin.H{"user": id})
}
