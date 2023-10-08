package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetLogin(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{"title": "login"})
}

func PostLogin(c *gin.Context) {
	c.JSON(http.StatusOK, "OK")
}
