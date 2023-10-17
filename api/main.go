package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"time"

	// TODO implement cookie sessions
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
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

	"github.com/solomonbaez/SB-Go-Newsletter-API/api/clients"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/configs"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/handlers"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/idempotency"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/routes"
	adminRoutes "github.com/solomonbaez/SB-Go-Newsletter-API/api/routes/admin"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/workers"
)

type App struct {
	database configs.DBSettings
	redis    configs.RedisSettings
	port     uint16
}

const baseUrl = "localhost"

// adjust for slowloris prevention
var readHeaderTimeout = 5 * time.Second

var app *App
var client *clients.SMTPClient

func init() {
	appCFG, e := configs.ConfigureApp()
	if e != nil {
		log.Fatal().
			Err(e).
			Msg("Failed to read database config")

		return
	}
	app = &App{
		appCFG.Database,
		appCFG.Redis,
		appCFG.Port,
	}

	cmd := flag.String("cfg", "", "")
	flag.Parse()

	client, e = clients.NewSMTPClient(cmd)
	if e != nil {
		log.Fatal().
			Err(e).
			Msg("Failed to create new SMTP Client")
	}
}

var pool *pgxpool.Pool
var enableTracing = false
var enableAuth = false

// server
func main() {
	parentContext := context.Background()

	if enableTracing {
		exporter, e := stdouttrace.New(stdouttrace.WithPrettyPrint())
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

	var e error
	// initialize database
	pool, e = initializeDatabase(parentContext)
	if e != nil {
		log.Fatal().
			Err(e).
			Msg("Failed to connect to postgres")

		return
	}
	defer pool.Close()

	log.Info().
		Msg("Connected to postgres")

	// initialize server components
	dh := handlers.NewDatabaseHandler(pool)
	dh.Context = parentContext

	go workers.WorkerLoop(parentContext, dh, client)

	router, listener, e := initializeServer(dh)
	if e != nil {
		log.Fatal().
			Err(e).
			Msg("Could not initialize server")

		return
	}
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)
	log.Info().
		Int("port", addr.Port).
		Msg(fmt.Sprintf("Listening: http://%v:%d", baseUrl, addr.Port))

	// server
	server := &http.Server{
		Handler:           router,
		ReadHeaderTimeout: readHeaderTimeout,
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
	pool, e := pgxpool.New(c, app.database.ConnectionString())
	if e != nil {
		return nil, e
	}
	if e = pool.Ping(context.Background()); e != nil {
		pool.Close()
		return nil, e
	}

	return pool, nil
}

func initializeServer(dh *handlers.DatabaseHandler) (*gin.Engine, net.Listener, error) {
	var e error
	// router
	router := gin.Default()

	key1, e := handlers.GenerateCSPRNG(32)
	if e != nil {
		return nil, nil, e
	}
	key2, e := handlers.GenerateCSPRNG(32)
	if e != nil {
		return nil, nil, e
	}
	// TODO migrate to cfg keys
	storeKeys := [][]byte{
		[]byte(key1),
		[]byte(key2),
	}

	redisStore, e := redis.NewStore(10, app.redis.Conn, app.redis.ConnectionString(), "", storeKeys[0], storeKeys[1])
	if e != nil {
		log.Fatal().Err(e).Msg("Failed to connect to redis")
	}
	router.Use(sessions.Sessions("admin", redisStore))

	// Get the absolute path to the "templates" directory
	templatesDir, e := filepath.Abs("./api/templates")
	if e != nil {
		log.Fatal().Err(e).Msg("Failed to get the absolute path to templates")
		return nil, nil, e
	}

	router.LoadHTMLGlob(filepath.Join(templatesDir, "*"))

	// custom middleware
	if enableTracing {
		router.Use(TraceMiddleware())
	}
	// disable during dev
	if enableAuth {
		var users gin.Accounts
		router.Use(func(c *gin.Context) {
			users, e = adminRoutes.GetUsers(c, dh)
			if e != nil {
				log.Fatal().
					Err(e).
					Msg("Failed to enable BasicAuth")
				return
			}
			gin.BasicAuth(users)
		})
	}

	// define admin group
	admin := router.Group("/admin")
	admin.Use(AdminMiddleware())
	admin.GET("/dashboard", adminRoutes.GetAdminDashboard)
	admin.GET("/password", adminRoutes.GetChangePassword)
	admin.POST("/password", func(c *gin.Context) { adminRoutes.PostChangePassword(c, dh) })
	admin.GET("/logout", adminRoutes.Logout)
	admin.GET("/subscribers", func(c *gin.Context) { adminRoutes.GetSubscribers(c, dh) })
	admin.GET("/subscribers/:id", func(c *gin.Context) { adminRoutes.GetSubscriberByID(c, dh) })
	admin.GET("/newsletter", adminRoutes.GetNewsletter)
	admin.POST("/newsletter", func(c *gin.Context) { adminRoutes.PostNewsletter(c, dh, client) })

	admin.GET("/responses", func(c *gin.Context) { idempotency.GetSavedResponses(c, dh) })

	router.GET("/health", handlers.HealthCheck)
	router.GET("/home", routes.Home)
	router.GET("/login", routes.GetLogin)
	router.POST("/login", func(c *gin.Context) { routes.PostLogin(c, dh) })
	router.POST("/subscribe", func(c *gin.Context) { routes.Subscribe(c, dh, client) })
	router.GET("/confirm/:token", func(c *gin.Context) { routes.ConfirmSubscriber(c, dh) })

	// listener
	listener, e := net.Listen("tcp", fmt.Sprintf("localhost:%v", app.port))
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

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get("user")
		if user == nil {
			c.Header("X-Redirect", "Forbidden")
			c.Redirect(http.StatusSeeOther, "../login")
			c.Abort()
			return
		}

		c.Next()
	}
}
