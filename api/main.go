package main

import (
	"fmt"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/handlers"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/logger"
)

// set dev port
const dev_port = 8000

func main() {
	// router
	router := gin.Default()
	router.GET("/health", handlers.HealthCheck)
	router.GET("/subscribers", handlers.GetSubscribers)
	router.POST("/subscribe", handlers.Subscribe)

	// listener
	listener, e := net.Listen("tcp", fmt.Sprintf("localhost:%v", dev_port))
	if e != nil {

		listener, e = net.Listen("tcp", "localhost:0")
		if e != nil {
			logger.Fatal(e.Error() + " - could not bind listener")
			return
		}

	}

	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)
	logger.Info(fmt.Sprintf("Listening on port %d\n", addr.Port))

	// server
	server := &http.Server{
		Handler: router,
	}

	e = server.Serve(listener)
	if e != nil {
		logger.Fatal(e.Error() + " - could not start server")
		return
	}
}
