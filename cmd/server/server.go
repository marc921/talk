package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/sync/errgroup"

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

	controller := server.NewServerController(logger)
	websocketHub := server.NewWebSocketHub(logger)

	api := server.NewAPI(
		logger,
		authenticator,
		controller,
		websocketHub,
	)

	// Echo instance
	e := echo.New()

	e.AutoTLSManager.HostPolicy = autocert.HostWhitelist("marcbrun.eu")
	e.AutoTLSManager.Cache = autocert.DirCache("/var/www/.cache")
	e.Use(middleware.Logger())
	e.Debug = true

	// Client
	e.GET("/client", func(c echo.Context) error {
		return c.File("./public/talkclient")
	})

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

	websocket := v1.Group("/ws")
	websocket.Use(echojwt.JWT([]byte(config.AuthTokenSecretKey)))
	websocket.GET("/:username", api.RegisterWebsocketClient)

	// Start server
	errGrp, ctx := errgroup.WithContext(ctx)

	errGrp.Go(func() error {
		err := websocketHub.Run(ctx)
		if err != nil {
			return fmt.Errorf("websocketHub.Run: %w", err)
		}
		return nil
	})

	errGrp.Go(func() error {
		err := e.StartAutoTLS(":443")
		if err != nil {
			return fmt.Errorf("e.StartAutoTLS: %w", err)
		}
		return nil
	})

	errGrp.Go(func() error {
		<-ctx.Done()
		gracePeriod := time.Minute
		logger.Info("shutting down echo server", zap.Duration("grace_period", gracePeriod))
		shutdownCtx, cancel := context.WithTimeout(context.Background(), gracePeriod)
		defer cancel()
		err := e.Shutdown(shutdownCtx)
		if err != nil {
			return fmt.Errorf("server shutdown: %w", err)
		}
		return nil
	})

	err = errGrp.Wait()
	if err != nil {
		if errors.Is(err, context.Canceled) {
			logger.Info("shutting down")
			return
		}
		logger.Fatal("errGrp.Wait", zap.Error(err))
	}
}

func OnSignal(f func(), logger *zap.Logger) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigs
	logger.Info("signal received", zap.String("signal", sig.String()))
	f()
}
