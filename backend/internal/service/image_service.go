package service

import (
	"errors"

	"serverdock/internal/dto"
	"serverdock/internal/model"
	"serverdock/internal/repository"
)

type ImageService struct {
	imageRepo     *repository.ImageRepo
	serverService *ServerService
	dockerService DockerService
}

func NewImageService(imageRepo *repository.ImageRepo, serverService *ServerService, dockerService DockerService) *ImageService {
	return &ImageService{imageRepo: imageRepo, serverService: serverService, dockerService: dockerService}
}

func (s *ImageService) Create(req *dto.CreateImageRequest) (*dto.ImageResponse, error) {
	// Verify server exists
	_, err := s.serverService.GetByID(req.ServerID)
	if err != nil {
		return nil, errors.New("server not found")
	}

	image := &model.Image{
		ServerID:      req.ServerID,
		DockerImageID: req.ImageID,
		Name:          req.Name,
		ImageAddress:  req.ImageAddress,
	}

	if err := s.imageRepo.Create(image); err != nil {
		return nil, err
	}
	return s.toResponse(image), nil
}

func (s *ImageService) GetByID(id uint) (*dto.ImageResponse, error) {
	image, err := s.imageRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("image not found")
	}
	return s.toResponse(image), nil
}

func (s *ImageService) List(serverID *uint) ([]dto.ImageResponse, error) {
	images, err := s.imageRepo.List(serverID)
	if err != nil {
		return nil, err
	}
	var responses []dto.ImageResponse
	for _, img := range images {
		responses = append(responses, *s.toResponse(&img))
	}
	return responses, nil
}

func (s *ImageService) Update(id uint, req *dto.UpdateImageRequest) (*dto.ImageResponse, error) {
	image, err := s.imageRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("image not found")
	}

	if req.Name != "" {
		image.Name = req.Name
	}
	if req.ImageAddress != "" {
		image.ImageAddress = req.ImageAddress
	}

	if err := s.imageRepo.Update(image); err != nil {
		return nil, err
	}
	return s.toResponse(image), nil
}

func (s *ImageService) Delete(id uint) error {
	_, err := s.imageRepo.FindByID(id)
	if err != nil {
		return errors.New("image not found")
	}
	return s.imageRepo.Delete(id)
}

// ListRemoteImages lists Docker images on a remote server.
func (s *ImageService) ListRemoteImages(serverID uint) ([]dto.RemoteImage, error) {
	server, cred, err := s.serverService.ResolveServer(serverID)
	if err != nil {
		return nil, err
	}

	return s.dockerService.ListImages(server.Hostname, server.Port, server.User, server.AuthType, cred)
}

// PullRemoteImage pulls an image on a remote server.
func (s *ImageService) PullRemoteImage(serverID uint, image string) (string, error) {
	server, cred, err := s.serverService.ResolveServer(serverID)
	if err != nil {
		return "", err
	}

	return s.dockerService.PullImage(server.Hostname, server.Port, server.User, server.AuthType, cred, image)
}

// RemoveRemoteImage removes a Docker image from a remote server.
// Returns error if the image is referenced by a DB record.
func (s *ImageService) RemoveRemoteImage(serverID uint, imageID string) error {
	// Check if referenced by DB record
	_, err := s.imageRepo.FindByImageIDAndServerID(imageID, serverID)
	if err == nil {
		return errors.New("cannot delete: image is registered as available in the system")
	}

	server, cred, err := s.serverService.ResolveServer(serverID)
	if err != nil {
		return err
	}

	return s.dockerService.RemoveImage(server.Hostname, server.Port, server.User, server.AuthType, cred, imageID)
}

func (s *ImageService) toResponse(image *model.Image) *dto.ImageResponse {
	return &dto.ImageResponse{
		ID:           image.ID,
		ServerID:     image.ServerID,
		ImageID:      image.DockerImageID,
		Name:         image.Name,
		ImageAddress: image.ImageAddress,
		CreatedAt:    image.CreatedAt,
	}
}
