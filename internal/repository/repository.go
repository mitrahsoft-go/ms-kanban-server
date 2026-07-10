package repository

import "gorm.io/gorm"

type Repository interface {
}

func InitRepository(db *gorm.DB) Repository {
	return &repository{
		DB: db,
	}
}

type repository struct {
	DB *gorm.DB
}
