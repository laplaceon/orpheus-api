package api

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	cache "github.com/chenyahui/gin-cache"
	"github.com/chenyahui/gin-cache/persist"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
)

func Server() {
	// read command-line flags
	host := flag.String("host", "localhost", "Server host")
	port := flag.Int("port", 8080, "Server port")
	docker := flag.Bool("docker", false, "Running in docker")
	flag.Parse()

	// prepare service, http handler and server
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	service := InitService()

	traceRate := 1.0
	if gin.Mode() == gin.ReleaseMode {
		traceRate = 0.2
	}

	if err := sentry.Init(sentry.ClientOptions{
		Dsn:           "https://abd5e43ad02ef85f85e6756351bff329@o4505818349240320.ingest.sentry.io/4505818351534080",
		EnableTracing: true,
		// Set TracesSampleRate to 1.0 to capture 100%
		// of transactions for performance monitoring.
		// We recommend adjusting this value in production,
		TracesSampleRate: traceRate,
	}); err != nil {
		fmt.Printf("Sentry initialization failed: %v", err)
	}

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, "Authorization")
	corsConfig.AllowAllOrigins = true

	router.Use(sentrygin.New(sentrygin.Options{}), cors.New(corsConfig))

	defer service.db.Close()

	memoryStore := persist.NewMemoryStore(6 * time.Hour)

	// apis
	api := router.Group("/v1")
	api.POST("users", service.CreateUser)
	api.POST("users/auth", service.GetUser)
	api.GET("actions", cache.CacheByRequestURI(memoryStore, 1*time.Hour), service.GetActions)
	api.POST("payments", service.ProcessPaymentFromStripe)

	api.Use(AuthRequired)
	{
		api.GET("users/:id", service.GetUserWithId)
		api.GET("users/:id/history", service.GetAllHistory)
		api.GET("history/:id", service.GetHistoryItem)
		api.GET("history/:id/generated", service.GetGeneratedFromHistory)
		api.POST("actions/genretransfer", service.CreateGenreTransferRequest)
	}

	// serve static files
	// router.Use(static.Serve("/", static.LocalFile("./build", true)))
	router.NoRoute(func(c *gin.Context) { // fallback
		c.File("./build/index.html")
	})

	var serverPath string
	if *docker {
		serverPath = "0.0.0.0:8080"
		log.Println("Server started at http://localhost:8080 ...")
	} else {
		serverPath = fmt.Sprintf("%s:%d", *host, *port)
		log.Printf("Server started at http://%s ...\n", serverPath)
	}

	server := &http.Server{
		Addr:         serverPath,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// start server
	go func() {
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalln(err)
		}
	}()

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server ...")

	service.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalln(err)
	}
	log.Println("Server exiting")
}
