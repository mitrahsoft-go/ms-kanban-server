package services

import (
	"github.com/ms-kanban-server/internal/repository"
	"go.uber.org/zap"
)

type Service interface {
}

func InitService(repo repository.Repository, logger *zap.Logger) Service {
	return &service{
		Repo:   repo,
		logger: logger,
	}
}

type service struct {
	Repo   repository.Repository
	logger *zap.Logger
}
