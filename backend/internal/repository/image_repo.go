package repository

import (
	"serverdock/internal/model"

	"gorm.io/gorm"
)

type ImageRepo struct {
	db *gorm.DB
}

func NewImageRepo(db *gorm.DB) *ImageRepo {
	return &ImageRepo{db: db}
}

func (r *ImageRepo) Create(image *model.Image) error {
	return r.db.Create(image).Error
}

func (r *ImageRepo) FindByID(id uint) (*model.Image, error) {
	var image model.Image
	err := r.db.Preload("Server").First(&image, id).Error
	if err != nil {
		return nil, err
	}
	return &image, nil
}

func (r *ImageRepo) List(serverID *uint) ([]model.Image, error) {
	var images []model.Image
	q := r.db.Preload("Server").Order("id desc")
	if serverID != nil {
		q = q.Where("server_id = ?", *serverID)
	}
	err := q.Find(&images).Error
	return images, err
}

func (r *ImageRepo) Update(image *model.Image) error {
	return r.db.Save(image).Error
}

func (r *ImageRepo) Delete(id uint) error {
	return r.db.Delete(&model.Image{}, id).Error
}

func (r *ImageRepo) FindByImageIDAndServerID(imageID string, serverID uint) (*model.Image, error) {
	var image model.Image
	err := r.db.Where("image_id = ? AND server_id = ?", imageID, serverID).First(&image).Error
	if err != nil {
		return nil, err
	}
	return &image, nil
}
