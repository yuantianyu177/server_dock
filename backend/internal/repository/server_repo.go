package repository

import (
	"serverdock/internal/model"

	"gorm.io/gorm"
)

type ServerRepo struct {
	db *gorm.DB
}

func NewServerRepo(db *gorm.DB) *ServerRepo {
	return &ServerRepo{db: db}
}

func (r *ServerRepo) Create(server *model.Server) error {
	return r.db.Create(server).Error
}

func (r *ServerRepo) FindByID(id uint) (*model.Server, error) {
	var server model.Server
	err := r.db.First(&server, id).Error
	if err != nil {
		return nil, err
	}
	return &server, nil
}

func (r *ServerRepo) List() ([]model.Server, error) {
	var servers []model.Server
	err := r.db.Order("id desc").Find(&servers).Error
	return servers, err
}

func (r *ServerRepo) Update(server *model.Server) error {
	return r.db.Save(server).Error
}

func (r *ServerRepo) Delete(id uint) error {
	return r.db.Delete(&model.Server{}, id).Error
}
