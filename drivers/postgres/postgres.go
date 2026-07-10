package postgres

import (
	"fmt"

	"github.com/ms-kanban-server/config/configs"
	"github.com/ms-kanban-server/drivers/migration"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB(config *configs.Config) (*gorm.DB, error) {

	// Create the connection string using the configuration values
	connectionKey := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", config.Database.Host, config.Database.Port, config.Database.Username, config.Database.Password, config.Database.Name, config.Database.SSLMode)

	// Initialize the database connection and perform any necessary setup here
	dbConn, err := gorm.Open(postgres.Open(connectionKey), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("Failed connecting to DB : %v", err)
	}

	err = migration.AutoMigrate(dbConn)
	if err != nil {
		return nil, fmt.Errorf("Failed to migrate : %v", err)
	}

	return dbConn, nil
}
