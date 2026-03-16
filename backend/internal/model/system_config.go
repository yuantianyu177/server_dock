package model

type SystemConfig struct {
	Key   string `gorm:"primaryKey"`
	Value string `gorm:"not null"`
}
