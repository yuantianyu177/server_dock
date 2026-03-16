package service

import (
	"serverdock/internal/dto"
	"serverdock/internal/repository"
	"strconv"
)

var defaultConfigs = map[string]string{
	"port_range_start":           "20000",
	"port_range_end":             "30000",
	"extra_ports_per_container":  "5",
	"default_volume_mount_path":  "/data",
	"docker_extra_args":          "",
	"email_enabled":              "false",
	"admin_email":                "",
	"smtp_host":                  "",
	"smtp_port":                  "587",
	"smtp_username":              "",
	"smtp_password":              "",
	"smtp_use_tls":               "true",
}

type ConfigService struct {
	configRepo *repository.ConfigRepo
}

func NewConfigService(configRepo *repository.ConfigRepo) *ConfigService {
	return &ConfigService{configRepo: configRepo}
}

// EnsureDefaults initializes default config values if they don't exist.
func (s *ConfigService) EnsureDefaults() error {
	for key, val := range defaultConfigs {
		if _, err := s.configRepo.Get(key); err != nil {
			if err := s.configRepo.Set(key, val); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *ConfigService) Get(key string) string {
	val, err := s.configRepo.Get(key)
	if err != nil {
		if def, ok := defaultConfigs[key]; ok {
			return def
		}
		return ""
	}
	return val
}

func (s *ConfigService) GetInt(key string) int {
	val := s.Get(key)
	n, _ := strconv.Atoi(val)
	return n
}

func (s *ConfigService) Set(key, value string) error {
	return s.configRepo.Set(key, value)
}

func (s *ConfigService) List() ([]dto.ConfigItem, error) {
	configs, err := s.configRepo.GetAll()
	if err != nil {
		return nil, err
	}

	var items []dto.ConfigItem
	for _, cfg := range configs {
		items = append(items, dto.ConfigItem{Key: cfg.Key, Value: cfg.Value})
	}
	return items, nil
}

func (s *ConfigService) GetAllAsMap() (map[string]string, error) {
	configs, err := s.configRepo.GetAll()
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	// Start with defaults
	for k, v := range defaultConfigs {
		result[k] = v
	}
	// Override with DB values
	for _, cfg := range configs {
		result[cfg.Key] = cfg.Value
	}
	return result, nil
}
