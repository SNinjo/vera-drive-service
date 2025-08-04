package db

import (
	"vera-identity-service/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewDatabase(config *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(config.DatabaseURL), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}
