package repository

import (
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Repository interface {
}

func InitRepository(db *gorm.DB, logger *zap.Logger) Repository {
	return &repository{
		DB:     db,
		logger: logger,
	}
}

type repository struct {
	DB     *gorm.DB
	logger *zap.Logger
}
