package model

import "time"

type Application struct {
	ID             uint   `gorm:"primaryKey"`
	ApplicantName  string `gorm:"not null"`
	ApplicantEmail string `gorm:"not null"`
	ServerID       uint   `gorm:"not null;index"`
	ImageID        uint   `gorm:"not null;index"`
	Status         string `gorm:"not null;default:pending"` // pending, approved, rejected
	AdminNotes     string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Server         Server `gorm:"foreignKey:ServerID;references:ID"`
	Image          Image  `gorm:"foreignKey:ImageID;references:ID"`
}
