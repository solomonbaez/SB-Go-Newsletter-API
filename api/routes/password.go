package routes

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func GetChangePassword(c *gin.Context) {
	session := sessions.Default(c)
	flashes := session.Flashes()

	c.HTML(http.StatusOK, "password.html", gin.H{"flashes": flashes})
}
