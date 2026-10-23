package db

import (
	"fmt"
	"os"

	"example.com/demo/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectPostgres() (*gorm.DB, error) {
	dsn := os.Getenv("DATABASE_URL")

	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set")
	}

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}
	if err := database.AutoMigrate(&model.Room{}); err != nil {
		return nil, fmt.Errorf("migrate room table: %w", err)
	}
	return database, nil
}

func ClosePostgres(database *gorm.DB) error {
	sqlDB, err := database.DB()
	if err != nil {
		return fmt.Errorf("get postgres connection: %w", err)
	}
	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("close postgres connection: %w", err)
	}
	return nil
}
