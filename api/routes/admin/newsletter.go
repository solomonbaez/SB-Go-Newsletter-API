package routes

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/idempotency"
)

func GetNewsletter(c *gin.Context) {
	session := sessions.Default(c)
	flashes := session.Flashes()

	key, e := idempotency.GenerateIdempotencyKey()
	if e != nil {
		flashes = append(flashes, "Failed to generate idempotency key, please reload session")
	}
	session.Save()

	c.HTML(http.StatusOK, "newsletter.html", gin.H{"flashes": flashes, "idempotency_key": key})
}
