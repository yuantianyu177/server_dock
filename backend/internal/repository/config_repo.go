package repository

import (
	"serverdock/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ConfigRepo struct {
	db *gorm.DB
}

func NewConfigRepo(db *gorm.DB) *ConfigRepo {
	return &ConfigRepo{db: db}
}

func (r *ConfigRepo) Get(key string) (string, error) {
	var cfg model.SystemConfig
	err := r.db.Where("key = ?", key).First(&cfg).Error
	if err != nil {
		return "", err
	}
	return cfg.Value, nil
}

func (r *ConfigRepo) Set(key, value string) error {
	cfg := model.SystemConfig{Key: key, Value: value}
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value"}),
	}).Create(&cfg).Error
}

func (r *ConfigRepo) GetAll() ([]model.SystemConfig, error) {
	var configs []model.SystemConfig
	err := r.db.Order("key").Find(&configs).Error
	return configs, err
}
