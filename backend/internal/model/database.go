package model

import (
	"log/slog"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func InitDB(path string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&Admin{}, &Server{}, &Image{}, &Application{}, &SystemConfig{}); err != nil {
		return nil, err
	}
	slog.Info("Database initialized", "path", path)
	return db, nil
}
