package main

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/configs"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/handlers"
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

var db *pgx.Conn

// server
func main() {
	// initialize database
	var e error
	db, e = intialize_database(context.Background())
	if e != nil {
		log.Fatal().
			Err(e).
			Msg("Failed to connect to postgres")

		return
	}

	log.Info().
		Msg("Connected to postgres")

	defer db.Close(context.Background())

	// initialize server components
	router, listener, e := initialize_server()
	if e != nil {
		log.Fatal().
			Err(e).
			Msg("Could not initialize server")
	}

	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)
	log.Info().
		Int("port", addr.Port).
		Msg(fmt.Sprintf("Listening: http://%v:%d", "localhost", addr.Port))

	// server
	server := &http.Server{
		Handler: router,
	}

	e = server.Serve(listener)
	if e != nil {
		log.Fatal().
			Err(e).
			Msg("Could not start server")

		return
	}
}

func intialize_database(c context.Context) (*pgx.Conn, error) {
	db, e := pgx.Connect(c, cfg.Database.Connection_String())
	if e != nil {
		return nil, e
	}
	if e = db.Ping(context.Background()); e != nil {
		db.Close(c)
		return nil, e
	}

	return db, nil
}

func initialize_server() (*gin.Engine, net.Listener, error) {
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
				Err(e).
				Msg("Could not bind listener")

			return nil, nil, e
		}
	}

	return router, listener, nil
}
