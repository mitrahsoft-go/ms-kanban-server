package migration

import "gorm.io/gorm"

func AutoMigrate(dbConn *gorm.DB) error {
	// Perform auto-migration for your models here
	// Example: dbConn.AutoMigrate(&YourModel{})
	return nil
}
