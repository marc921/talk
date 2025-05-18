package main

import (
	"context"
	"errors"
	"fmt"
	"path"

	"go.uber.org/zap"

	"github.com/marc921/talk/internal/client"
	"github.com/marc921/talk/internal/client/database"
	"github.com/marc921/talk/internal/types/openapi"
	"github.com/spf13/cobra"
)

var (
	fileMode   bool
	outputFile string
)

var rootCmd = &cobra.Command{
	Use:   "client",
	Short: "Talk client TUI",
	Run: func(cmd *cobra.Command, args []string) {
		// Default: launch UI
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

		db, err := database.GetOrCreateSQLite3DB(path.Join(config.HomeDir, "database.sqlite3"))
		if err != nil {
			logger.Fatal("database.GetOrCreateSQLite3DB", zap.Error(err))
		}

		openapiClient, err := openapi.NewClientWithResponses(
			config.Server.URL,
		)
		if err != nil {
			logger.Fatal("openapi.NewClientWithResponses", zap.Error(err))
		}

		err = client.InitUI(config, openapiClient, db)
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
	},
}

func mustGetCLIHandler(
	ctx context.Context,
) *client.CLIHandler {
	logger, err := zap.NewDevelopment()
	if err != nil {
		zap.L().Fatal("zap.NewDevelopment", zap.Error(err))
	}

	config, err := client.LoadConfig(ctx)
	if err != nil {
		if errors.Is(err, client.ErrAbortedByUser) {
			return nil
		}
		logger.Fatal("LoadConfig", zap.Error(err))
	}

	db, err := database.GetOrCreateSQLite3DB(path.Join(config.HomeDir, "database.sqlite3"))
	if err != nil {
		logger.Fatal("database.GetOrCreateSQLite3DB", zap.Error(err))
	}

	openapiClient, err := openapi.NewClientWithResponses(
		config.Server.URL,
	)
	if err != nil {
		logger.Fatal("openapi.NewClientWithResponses", zap.Error(err))
	}

	controller := client.NewController(openapiClient, db)
	return client.NewCLIHandler(logger, controller)
}

var messageCmd = &cobra.Command{
	Use:   "message",
	Short: "Messages commands",
}

var messageReadCmd = &cobra.Command{
	Use:  "read [-o] <username>",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithCancel(cmd.Context())
		defer cancel()
		cliHandler := mustGetCLIHandler(ctx)

		username := args[0]
		if outputFile != "" {
			// Use outputFile as directory for message files
			return cliHandler.ReadMessagesToDir(ctx, username, outputFile)
		}
		return cliHandler.ReadMessages(ctx, username)
	},
}

var messageSendCmd = &cobra.Command{
	Use:   "send [--file] <sender> <recipient> <message|file_path>",
	Short: "Send a message or file",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithCancel(cmd.Context())
		defer cancel()
		cliHandler := mustGetCLIHandler(ctx)

		sender := args[0]
		recipient := args[1]
		if fileMode {
			filePath := args[2]
			err := cliHandler.SendFile(ctx, sender, recipient, filePath)
			if err != nil {
				return fmt.Errorf("cliHandler.SendFile: %w", err)
			}
		} else {
			message := args[2]
			err := cliHandler.SendMessage(ctx, sender, recipient, message)
			if err != nil {
				return fmt.Errorf("cliHandler.SendMessage: %w", err)
			}
		}
		return nil
	},
}

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "User commands",
}

var createUserCmd = &cobra.Command{
	Short: "Create a new user",
	Use:   "create <username>",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithCancel(cmd.Context())
		defer cancel()
		cliHandler := mustGetCLIHandler(ctx)

		username := args[0]
		err := cliHandler.CreateUser(ctx, username)
		if err != nil {
			return fmt.Errorf("cliHandler.CreateUser: %w", err)
		}
		return nil
	},
}

func main() {
	messageSendCmd.Flags().BoolVar(&fileMode, "file", false, "Send a file instead of a text message")
	messageReadCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Write output to file instead of stdout")
	messageCmd.AddCommand(messageSendCmd)
	messageCmd.AddCommand(messageReadCmd)
	rootCmd.AddCommand(messageCmd)

	userCmd.AddCommand(createUserCmd)
	rootCmd.AddCommand(userCmd)

	_ = rootCmd.Execute()
}
