package model

import "time"

type Image struct {
	ID            uint   `gorm:"primaryKey"`
	ServerID      uint   `gorm:"not null;index"`
	DockerImageID string `gorm:"column:image_id;type:text;not null"`
	Name          string `gorm:"not null"`
	ImageAddress  string `gorm:"not null"`
	CreatedAt     time.Time
	Server        Server `gorm:"foreignKey:ServerID"`
}
