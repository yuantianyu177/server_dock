package service

import (
	"serverdock/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var defaultConfigs = map[string]string{
	"port_range_start": "20000", "port_range_end": "30000", "extra_ports_per_container": "5",
	"default_volume_mount_path": "/data", "docker_extra_args": "", "email_enabled": "false",
	"admin_email": "", "smtp_host": "", "smtp_port": "587", "smtp_username": "",
	"smtp_password": "", "smtp_use_tls": "true",
}

type ConfigService struct{ db *gorm.DB }

func NewConfigService(db *gorm.DB) *ConfigService { return &ConfigService{db: db} }

func (s *ConfigService) EnsureDefaults() error {
	for key, value := range defaultConfigs {
		if err := s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&model.SystemConfig{Key: key, Value: value}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *ConfigService) Get(key string) string {
	var config model.SystemConfig
	if err := s.db.First(&config, "key = ?", key).Error; err == nil {
		return config.Value
	}
	return defaultConfigs[key]
}

func (s *ConfigService) Set(key, value string) error {
	config := model.SystemConfig{Key: key, Value: value}
	return s.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "key"}}, DoUpdates: clause.AssignmentColumns([]string{"value"}),
	}).Create(&config).Error
}

func (s *ConfigService) GetAllAsMap() (map[string]string, error) {
	var configs []model.SystemConfig
	if err := s.db.Find(&configs).Error; err != nil {
		return nil, err
	}
	result := make(map[string]string, len(defaultConfigs))
	for key, value := range defaultConfigs {
		result[key] = value
	}
	for _, config := range configs {
		result[config.Key] = config.Value
	}
	return result, nil
}
