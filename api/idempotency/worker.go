package idempotency

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/clients"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/handlers"
)

func DeliveryWorker(c *gin.Context, dh *handlers.DatabaseHandler, client *clients.SMTPClient) {
	var e error
	go func() {
		if e = TryExecuteTask(c, dh, client); e != nil {
			if e.Error() == "No rows in result set" {
				time.Sleep(10 * time.Second)
			} else {
				time.Sleep(time.Second)
			}
		}
	}()
}
