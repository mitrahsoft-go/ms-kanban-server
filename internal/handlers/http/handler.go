package handlers

import (
	"github.com/ms-kanban-server/internal/services"
	"go.uber.org/zap"
)

func InitHandler(service services.Service, logger *zap.Logger) *handler {
	return &handler{
		service: service,
		logger:  logger,
	}
}

type handler struct {
	service services.Service
	logger  *zap.Logger
}
