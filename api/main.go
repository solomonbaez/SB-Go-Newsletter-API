package main

import (
	"fmt"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/configs"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/handlers"
	// "github.com/solomonbaez/SB-Go-Newsletter-API/api/logger"
	"github.com/rs/zerolog/log"
)

// generate application settings
var cfg configs.AppSettings

func init() {
	database, port := configs.ConfigureApp()
	cfg = configs.AppSettings{
		Database: database,
		Port:     port,
	}
}

// server
func main() {
	// router
	router := gin.Default()
	router.GET("/health", handlers.HealthCheck)
	router.GET("/subscribers", handlers.GetSubscribers)
	router.POST("/subscribe", handlers.Subscribe)

	// listener
	listener, e := net.Listen("tcp", fmt.Sprintf("localhost:%v", cfg.Port))
	if e != nil {

		listener, e = net.Listen("tcp", "localhost:0")
		if e != nil {
			log.Fatal().
				Msg(fmt.Sprintf("ERROR: %v - could not bind listener", e.Error()))

			return
		}

	}

	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)
	log.Info().
		Msg(fmt.Sprintf("Listening on port %d\n", addr.Port))

	// server
	server := &http.Server{
		Handler: router,
	}

	e = server.Serve(listener)
	if e != nil {
		log.Fatal().
			Msg(fmt.Sprintf("ERROR: %v - could not start server", e.Error()))
		return
	}
}
