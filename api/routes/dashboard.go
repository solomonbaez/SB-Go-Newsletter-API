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
	flashes := session.Flashes()

	c.HTML(http.StatusOK, "dashboard.html", gin.H{"flashes": flashes})
}
