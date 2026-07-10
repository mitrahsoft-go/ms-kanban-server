package postgres

import (
	"fmt"

	"github.com/ms-kanban-server/config/configs"
	"github.com/ms-kanban-server/drivers/migration"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB(config *configs.Config) (*gorm.DB, error) {
	// Initialize the database connection and perform any necessary setup here

	connectionString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", config.Database.Host, config.Database.Port, config.Database.Username, config.Database.Password, config.Database.Name, config.Database.SSLMode)

	dbConn, err := gorm.Open(postgres.Open(connectionString), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the database: %w", err)
	}

	// Perform auto-migration for your models here
	if config.Database.AutoMigrate == "true" {
		err = migration.AutoMigrate(dbConn)
		if err != nil {
			return nil, fmt.Errorf("failed to auto-migrate the database: %w", err)
		}
	}

	return dbConn, nil
}
