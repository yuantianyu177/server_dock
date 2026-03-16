package repository

import (
	"serverdock/internal/model"

	"gorm.io/gorm"
)

type ApplicationRepo struct {
	db *gorm.DB
}

func NewApplicationRepo(db *gorm.DB) *ApplicationRepo {
	return &ApplicationRepo{db: db}
}

func (r *ApplicationRepo) Create(app *model.Application) error {
	return r.db.Create(app).Error
}

func (r *ApplicationRepo) FindByID(id uint) (*model.Application, error) {
	var app model.Application
	err := r.db.Preload("Server").Preload("Image").First(&app, id).Error
	if err != nil {
		return nil, err
	}
	return &app, nil
}

func (r *ApplicationRepo) List(status string) ([]model.Application, error) {
	var apps []model.Application
	q := r.db.Preload("Server").Preload("Image").Order("id desc")
	if status != "" {
		q = q.Where("status = ?", status)
	}
	err := q.Find(&apps).Error
	return apps, err
}

func (r *ApplicationRepo) Update(app *model.Application) error {
	return r.db.Save(app).Error
}
