package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv/v1.18.0"

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

var db *pgxpool.Pool
var enableTracing = true

// server
func main() {
	var exporter *stdouttrace.Exporter
	var e error
	if enableTracing {
		exporter, e = stdouttrace.New(stdouttrace.WithPrettyPrint())
		if e != nil {
			log.Fatal().
				Err(e).
				Msg("Failed to initialize telemetry")

			return
		}

		tp := trace.NewTracerProvider(
			trace.WithSyncer(exporter),
			trace.WithSampler(trace.AlwaysSample()),
			trace.WithResource(resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String("newsletter"),
			)),
		)

		otel.SetTracerProvider(tp)
	}

	// initialize database
	db, e = initializeDatabase(context.Background())
	if e != nil {
		log.Fatal().
			Err(e).
			Msg("Failed to connect to postgres")

		return
	}

	log.Info().
		Msg("Connected to postgres")

	defer db.Close()

	// initialize server components
	rh := handlers.NewRouteHandler(db)

	router, listener, e := initializeServer(rh)
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
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	e = server.Serve(listener)
	if e != nil {
		log.Fatal().
			Err(e).
			Msg("Could not start server")

		return
	}
}

func initializeDatabase(c context.Context) (*pgxpool.Pool, error) {
	db, e := pgxpool.New(c, cfg.Database.ConnectionString())
	if e != nil {
		return nil, e
	}
	if e = db.Ping(context.Background()); e != nil {
		db.Close()
		return nil, e
	}

	return db, nil
}

func initializeServer(rh *handlers.RouteHandler) (*gin.Engine, net.Listener, error) {
	// router
	router := gin.Default()

	// dev
	if enableTracing {
		router.Use(TraceMiddleware())
	}

	router.GET("/health", handlers.HealthCheck)
	router.GET("/subscribers", rh.GetSubscribers)
	router.GET("/subscribers/:id", rh.GetSubscriberByID)
	router.POST("/subscribe", rh.Subscribe)

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

func TraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// request identification
		requestID := uuid.NewString()
		c.Set("requestID", requestID)
		log.Info().
			Str("requestID", requestID).
			Msg(fmt.Sprintf("New %v request...", c.Request.Method))

		// tracing
		spanCTX := otel.
			GetTextMapPropagator().
			Extract(
				c.Request.Context(),
				propagation.HeaderCarrier(c.Request.Header),
			)

		ctx, span := otel.Tracer("http-server").Start(spanCTX, c.Request.URL.Path)
		span.SetAttributes(attribute.String("requestID", requestID))
		defer span.End()

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
