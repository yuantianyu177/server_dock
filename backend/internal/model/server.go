package model

import "time"

type Server struct {
	ID          uint   `gorm:"primaryKey"`
	Host        string `gorm:"not null"`
	Hostname    string `gorm:"not null"`
	Port        int    `gorm:"default:22"`
	User        string `gorm:"not null"`
	AuthType    string `gorm:"not null"` // "password" or "key"
	Credential  string `gorm:"not null"` // encrypted
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
