package migration

import (
	"gorm.io/gorm"
)

func AutoMigrate(dbConn *gorm.DB) error {

	err := dbConn.AutoMigrate()
	if err != nil {
		return err
	}

	return nil
}
