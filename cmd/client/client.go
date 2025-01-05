package main

import (
	"context"
	"errors"
	"path"

	"go.uber.org/zap"

	"github.com/marc921/talk/internal/client"
	"github.com/marc921/talk/internal/client/database"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger, err := zap.NewDevelopment()
	if err != nil {
		zap.L().Fatal("zap.NewDevelopment", zap.Error(err))
	}

	config, err := client.LoadConfig(ctx)
	if err != nil {
		if errors.Is(err, client.ErrAbortedByUser) {
			return
		}
		logger.Fatal("LoadConfig", zap.Error(err))
	}

	db, err := database.NewSQLite3DB(path.Join(config.HomeDir, "database.sqlite3"))
	if err != nil {
		logger.Fatal("database.NewSQLite3DB", zap.Error(err))
	}

	err = client.InitUI(config, db)
	if err != nil {
		logger.Fatal("NewUI", zap.Error(err))
	}
	defer client.UISingleton.Quit()

	err = client.UISingleton.Run(ctx)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		logger.Fatal("client.UISingleton.Run", zap.Error(err))
	}
	logger.Info("client.UISingleton.Run", zap.String("reason", "context.Canceled"))
}
