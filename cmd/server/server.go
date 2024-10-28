package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
	"golang.org/x/crypto/acme/autocert"

	"github.com/marc921/talk/internal/server"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger, err := zap.NewDevelopment()
	if err != nil {
		zap.L().Fatal("zap.NewDevelopment", zap.Error(err))
	}

	go OnSignal(cancel, logger)

	config, err := LoadConfig(ctx)
	if err != nil {
		logger.Fatal("LoadConfig", zap.Error(err))
	}

	authenticator := server.NewAuthenticator(
		config.AuthChallengeSecretKey,
		config.AuthTokenSecretKey,
		64,
		5*time.Minute,
		time.Hour,
	)

	controller := server.NewServerController(
		logger.With(zap.String("component", "controller")),
	)

	api := server.NewAPI(
		logger.With(zap.String("component", "api")),
		authenticator,
		controller,
	)

	// Echo instance
	e := echo.New()

	e.AutoTLSManager.HostPolicy = autocert.HostWhitelist("marcbrun.eu")
	e.AutoTLSManager.Cache = autocert.DirCache("/var/www/.cache")
	e.Use(middleware.Logger())

	go func() {
		<-ctx.Done()
		gracePeriod := time.Minute
		logger.Info("shutting down echo server", zap.Duration("grace_period", gracePeriod))
		shutdownCtx, cancel := context.WithTimeout(context.Background(), gracePeriod)
		defer cancel()
		err := e.Shutdown(shutdownCtx)
		if err != nil {
			logger.Error("server shutdown", zap.Error(err))
		}
	}()

	// Routes
	v1 := e.Group("/api/v1")
	v1.GET("/auth/:username", api.GetAuth)
	v1.POST("/auth/:username", api.PostAuth)

	v1.GET("/users/:username", api.GetUser)
	v1.POST("/users", api.AddUser)

	messages := v1.Group("/messages")
	messages.Use(echojwt.JWT([]byte(config.AuthTokenSecretKey)))
	messages.POST("/:username", api.AddMessage)
	messages.GET("/:username", api.GetMessages)

	// Start server
	e.Logger.Fatal(e.StartAutoTLS(":443"))
}

func OnSignal(f func(), logger *zap.Logger) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigs
	logger.Info("signal received", zap.String("signal", sig.String()))
	f()
}
