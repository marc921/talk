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

	ui, err := client.NewUI(config, db)
	if err != nil {
		logger.Fatal("NewUI", zap.Error(err))
	}
	defer ui.Quit()

	err = ui.Run(ctx)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		logger.Fatal("ui.Run", zap.Error(err))
	}
}

// err = alice.SendMessage(
// 	ctx,
// 	types.PlainText("hello, bob"),
// 	openapi.Username("bob"),
// )
// if err != nil {
// 	logger.Fatal("alice.SendMessage", zap.Error(err))
// }

// messages, err := bob.FetchMessages(ctx)
// if err != nil {
// 	logger.Fatal("bob.FetchMessages", zap.Error(err))
// }
// for _, message := range messages {
// 	fmt.Printf("%s -> %s: %q\n", message.Sender, message.Recipient, string(message.Plaintext))
// }
