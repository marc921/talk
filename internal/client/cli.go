package client

import (
	"context"
	"fmt"
	"os"
	"path"

	"go.uber.org/zap"
)

type CLIHandler struct {
	logger     *zap.Logger
	controller *Controller
}

func NewCLIHandler(
	logger *zap.Logger,
	controller *Controller,
) *CLIHandler {
	return &CLIHandler{
		logger:     logger.With(zap.String("component", "cli")),
		controller: controller,
	}
}

func (h *CLIHandler) CreateUser(
	ctx context.Context,
	username string,
) error {
	h.logger.Info(
		"Creating user...",
		zap.String("username", username),
	)

	err := h.controller.CreateUser(ctx, username)
	if err != nil {
		return fmt.Errorf("CreateUser: %w", err)
	}
	h.logger.Info("User created successfully!")
	return nil
}

func (h *CLIHandler) SendFile(
	ctx context.Context,
	sender,
	recipient,
	filePath string,
) error {
	h.logger.Info(
		"Sending file...",
		zap.String("sender", sender),
		zap.String("recipient", recipient),
		zap.String("filePath", filePath),
	)

	// Read file
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("os.ReadFile: %w", err)
	}

	// Get sender's local user
	user, err := h.controller.GetUser(ctx, sender)
	if err != nil {
		return fmt.Errorf("GetUser: %w", err)
	}

	// Send file as message
	err = user.SendMessage(ctx, fileBytes, recipient)
	if err != nil {
		return fmt.Errorf("SendMessage: %w", err)
	}
	h.logger.Info("File sent successfully!")
	return nil
}

func (h *CLIHandler) SendMessage(
	ctx context.Context,
	sender, recipient, message string,
) error {
	h.logger.Info(
		"Sending message...",
		zap.String("sender", sender),
		zap.String("recipient", recipient),
		zap.String("message", message),
	)

	// Get sender's local user
	user, err := h.controller.GetUser(ctx, sender)
	if err != nil {
		return fmt.Errorf("GetUser: %w", err)
	}

	// Send message as plaintext
	err = user.SendMessage(ctx, []byte(message), recipient)
	if err != nil {
		return fmt.Errorf("SendMessage: %w", err)
	}
	h.logger.Info("Message sent successfully!")
	return nil
}

func (h *CLIHandler) ReadMessages(
	ctx context.Context,
	username string,
) error {
	h.logger.Info(
		"Reading messages...",
		zap.String("username", username),
	)

	// Get user's local user
	user, err := h.controller.GetUser(ctx, username)
	if err != nil {
		return fmt.Errorf("GetUser: %w", err)
	}

	// Read messages
	messages, err := user.FetchMessages(ctx)
	if err != nil {
		return fmt.Errorf("FetchMessages: %w", err)
	}
	h.logger.Info("Messages read successfully!")

	// Print messages
	for _, message := range messages {
		fmt.Printf(`From %q:
%s
`,
			message.Sender,
			message.Content,
		)
	}
	return nil
}

func (h *CLIHandler) ReadMessagesToDir(ctx context.Context, username, outputDir string) error {
	h.logger.Info(
		"Reading messages to directory...",
		zap.String("username", username),
		zap.String("outputDir", outputDir),
	)

	user, err := h.controller.GetUser(ctx, username)
	if err != nil {
		return fmt.Errorf("GetUser: %w", err)
	}

	messages, err := user.FetchMessages(ctx)
	if err != nil {
		return fmt.Errorf("FetchMessages: %w", err)
	}
	h.logger.Info("Messages read successfully!")

	for _, message := range messages {
		senderDir := path.Join(outputDir, string(message.Sender))
		err := os.MkdirAll(senderDir, 0o755)
		if err != nil {
			return fmt.Errorf("failed to create sender dir: %w", err)
		}
		filePath := path.Join(senderDir, fmt.Sprintf("%d", message.ID))
		err = os.WriteFile(filePath, message.Content, 0o644)
		if err != nil {
			return fmt.Errorf("failed to write message file: %w", err)
		}
	}
	return nil
}
