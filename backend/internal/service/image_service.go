package service

import (
	"errors"
	"fmt"

	"serverdock/internal/dto"
	"serverdock/internal/model"

	"gorm.io/gorm"
)

type ImageService struct {
	db            *gorm.DB
	serverService *ServerService
}

func NewImageService(db *gorm.DB, serverService *ServerService) *ImageService {
	return &ImageService{db: db, serverService: serverService}
}

func (s *ImageService) Create(req *dto.CreateImageRequest) (*dto.ImageResponse, error) {
	if _, err := s.serverService.GetByID(req.ServerID); err != nil {
		return nil, errors.New("server not found")
	}
	image := &model.Image{
		ServerID: req.ServerID, DockerImageID: req.ImageID, Name: req.Name, ImageAddress: req.ImageAddress,
	}
	if err := s.db.Create(image).Error; err != nil {
		return nil, err
	}
	return imageResponse(image), nil
}

func (s *ImageService) List(serverID *uint) ([]dto.ImageResponse, error) {
	var images []model.Image
	query := s.db.Order("id desc")
	if serverID != nil {
		query = query.Where("server_id = ?", *serverID)
	}
	if err := query.Find(&images).Error; err != nil {
		return nil, err
	}
	responses := make([]dto.ImageResponse, len(images))
	for i := range images {
		responses[i] = *imageResponse(&images[i])
	}
	return responses, nil
}

func (s *ImageService) Delete(id uint) error {
	result := s.db.Delete(&model.Image{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("image not found")
	}
	return nil
}

func (s *ImageService) ListRemoteImages(serverID uint) ([]dto.RemoteImage, error) {
	output, err := s.serverService.ExecuteCommand(serverID,
		"docker images --format '{{.Repository}}\\t{{.Tag}}\\t{{.ID}}\\t{{.Size}}\\t{{.CreatedSince}}'")
	if err != nil {
		return nil, err
	}
	return parseDockerImages(output), nil
}

func (s *ImageService) PullRemoteImage(serverID uint, image string) (string, error) {
	return s.serverService.ExecuteCommand(serverID, fmt.Sprintf("docker pull '%s'", image))
}

func (s *ImageService) RemoveRemoteImage(serverID uint, imageID string) error {
	var image model.Image
	err := s.db.Where("image_id = ? AND server_id = ?", imageID, serverID).First(&image).Error
	if err == nil {
		return errors.New("cannot delete: image is registered as available in the system")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	_, err = s.serverService.ExecuteCommand(serverID, fmt.Sprintf("docker rmi '%s'", imageID))
	return err
}

func imageResponse(image *model.Image) *dto.ImageResponse {
	return &dto.ImageResponse{
		ID: image.ID, ServerID: image.ServerID, ImageID: image.DockerImageID,
		Name: image.Name, ImageAddress: image.ImageAddress, CreatedAt: image.CreatedAt,
	}
}
