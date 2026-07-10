package logger

import (
	"fmt"

	"github.com/ms-kanban-server/config/configs"
	"go.uber.org/zap"
)

func InitLogger(config *configs.Config) (*zap.Logger, error) {

	var Log *zap.Logger
	var err error

	if config.Logger.Type == "production" {
		Log, err = zap.NewProduction()
	} else {
		Log, err = zap.NewDevelopment()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	return Log, nil
}
