package services

import "github.com/ms-kanban-server/internal/repository"

type Service interface {
}

func InitService(repo repository.Repository) Service {
	return &service{
		Repo: repo,
	}
}

type service struct {
	Repo repository.Repository
}
