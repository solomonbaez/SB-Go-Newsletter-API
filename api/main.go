package main

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/handlers"
)

// set dev port
const dev_port = 8000

func main() {
	// router
	router := gin.Default()
	router.GET("/health", handlers.HealthCheck)

	// listener
	listener, e := net.Listen("tcp", fmt.Sprintf("localhost:%v", dev_port))
	if e != nil {

		listener, e = net.Listen("tcp", "localhost:0")
		if e != nil {
			log.Fatal(e)
			return
		}

	}

	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)
	log.Printf("Listening on port %d\n", addr.Port)

	// server
	server := &http.Server{
		Handler: router,
	}

	e = server.Serve(listener)
	if e != nil {
		log.Fatal(e)
		return
	}

}
