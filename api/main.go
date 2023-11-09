package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"path/filepath"
	// "sync"
	"time"

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
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/routes"
	adminRoutes "github.com/solomonbaez/SB-Go-Newsletter-API/api/routes/admin"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/workers"
)

type App struct {
	database *configs.DBSettings
	redis    *configs.RedisSettings
	port     uint16
}

// TODO switch to cfg baseUrl
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

var enableTracing = true
var pool *pgxpool.Pool

func main() {
	parentContext := context.Background()

	if enableTracing {
		if e := initializeTracing(); e != nil {
			log.Fatal().
				Err(e).
				Msg("Failed to intialize tracing")
		}
	}

	var e error
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

	// var wg sync.WaitGroup
	// // initialize newsletter delivery workers
	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// }()
	// wg.Wait()
	// go workers.PruningWorker(parentContext, dh)

	go workers.DeliveryWorker(parentContext, dh, client)

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

	if e = server.Serve(listener); e != nil {
		log.Fatal().
			Err(e).
			Msg("Could not start server")

		return
	}
	go func() {
		fmt.Println(http.ListenAndServe("localhost:6060", nil))
	}()

}

func initializeTracing() (err error) {
	exporter, e := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if e != nil {
		err = fmt.Errorf("failed to initialize telemetry: %w", e)
		log.Fatal().
			Err(e).
			Msg("failed to initialize telemetry")

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
	return
}

func initializeDatabase(c context.Context) (pool *pgxpool.Pool, err error) {
	var e error
	pool, e = pgxpool.New(c, app.database.ConnectionString())
	if e != nil {
		err = fmt.Errorf("failed to retrieve a database pool: %w", e)
		return
	}
	if e = pool.Ping(context.Background()); e != nil {
		err = fmt.Errorf("unresponsive database pool: %w", e)
		pool.Close()
		return
	}

	return
}

func initializeServer(dh *handlers.DatabaseHandler) (router *gin.Engine, listener net.Listener, err error) {
	router = gin.Default()
	if enableTracing {
		router.Use(TraceMiddleware())
	}
	router.Use(CSPMiddleware())
	router.Use(SecurityHeadersMiddleware())

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
		err = fmt.Errorf("failed to connect to redis: %w", e)
		log.Fatal().
			Err(e).
			Msg("failed to connect to redis")

		return
	}
	router.Use(sessions.Sessions("admin", redisStore))

	// Get the absolute path to the "templates" directory
	templatesDir, e := filepath.Abs("./api/templates")
	if e != nil {
		log.Fatal().Err(e).Msg("Failed to get the absolute path to templates")
		return nil, nil, e
	}

	router.LoadHTMLGlob(filepath.Join(templatesDir, "*"))

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

	router.GET("/debug/pprof/:id", gin.WrapH(http.DefaultServeMux))
	router.GET("/health", handlers.HealthCheck)
	router.GET("/login", routes.GetLogin)
	router.POST("/login", func(c *gin.Context) { routes.PostLogin(c, dh) })
	router.POST("/subscribe", func(c *gin.Context) { routes.Subscribe(c, dh) })
	router.GET("/confirm/:token", func(c *gin.Context) { routes.ConfirmSubscriber(c, dh) })

	// listener
	listener, e = net.Listen("tcp", fmt.Sprintf("localhost:%v", app.port))
	if e != nil {
		listener, e = net.Listen("tcp", "localhost:0")
		if e != nil {
			err = fmt.Errorf("failed to bind listener: %w", e)
			log.Fatal().
				Err(e).
				Msg("failed to bind listener")

			return
		}
	}

	return
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

func CSPMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' https://localhost:8000; img-src 'self' data:; style-src 'self' 'unsafe-inline';")
		c.Next()
	}
}

func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set HSTS header (enforces HTTPS for a specified duration)
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// Set X-Content-Type-Options header (prevents MIME type sniffing)
		c.Header("X-Content-Type-Options", "nosniff")

		c.Next()
	}
}
