package service

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"serverdock/internal/dto"
	"serverdock/internal/model"

	"gorm.io/gorm"
)

var (
	ErrApplicationNotFound   = errors.New("application not found")
	ErrApplicationNotPending = errors.New("application is not pending")
	ErrContainerProvisioning = errors.New("container provisioning failed")
)

type ApplicationService struct {
	db               *gorm.DB
	serverService    *ServerService
	containerService *ContainerService
	configService    *ConfigService
	sendAsync        func(to, subject, html string)
}

func NewApplicationService(db *gorm.DB, serverService *ServerService, containerService *ContainerService, configService *ConfigService, sendAsync func(string, string, string)) *ApplicationService {
	return &ApplicationService{
		db: db, serverService: serverService, containerService: containerService,
		configService: configService, sendAsync: sendAsync,
	}
}

func (s *ApplicationService) Submit(req *dto.SubmitApplicationRequest) (*dto.ApplicationResponse, error) {
	server, err := s.serverService.GetByID(req.ServerID)
	if err != nil {
		return nil, errors.New("server not found")
	}
	var image model.Image
	if err := s.db.First(&image, req.ImageID).Error; err != nil {
		return nil, errors.New("image not found")
	}
	if image.ServerID != req.ServerID {
		return nil, errors.New("image does not belong to the selected server")
	}

	application := &model.Application{
		ApplicantName: req.ApplicantName, ApplicantEmail: req.ApplicantEmail,
		ServerID: req.ServerID, ImageID: req.ImageID, Status: "pending",
	}
	if err := s.db.Create(application).Error; err != nil {
		return nil, err
	}
	application.Server.Host = server.Host
	application.Image = image
	if to := s.configService.Get("admin_email"); to != "" {
		s.notify(to, renderNewApplicationEmail(application.ApplicantName, application.ApplicantEmail, server.Host, image.Name))
	}
	return applicationResponse(application), nil
}

func (s *ApplicationService) List() ([]dto.ApplicationResponse, error) {
	var responses []dto.ApplicationResponse
	err := s.db.Table("applications").
		Select("applications.*, servers.host AS server_host, images.name AS image_name").
		Joins("LEFT JOIN servers ON servers.id = applications.server_id").
		Joins("LEFT JOIN images ON images.id = applications.image_id").
		Order("applications.id DESC").Scan(&responses).Error
	return responses, err
}

func (s *ApplicationService) Approve(id uint, adminNotes string) (*dto.ApplicationResponse, error) {
	application, err := s.find(id)
	if err != nil {
		return nil, ErrApplicationNotFound
	}
	if application.Status != "pending" {
		return nil, ErrApplicationNotPending
	}
	imageAddress := strings.TrimSpace(application.Image.ImageAddress)
	if imageAddress == "" {
		return nil, fmt.Errorf("%w: image %d has empty image address", ErrContainerProvisioning, application.ImageID)
	}

	config, err := s.configService.GetAllAsMap()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	portStart, _ := strconv.Atoi(config["port_range_start"])
	portEnd, _ := strconv.Atoi(config["port_range_end"])
	extraPorts, _ := strconv.Atoi(config["extra_ports_per_container"])
	result, err := s.containerService.CreateContainer(
		application.ServerID, fmt.Sprintf("container-%d", application.ID), imageAddress,
		config["docker_extra_args"], portStart, portEnd, extraPorts, config["default_volume_mount_path"],
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrContainerProvisioning, err)
	}

	application.Status, application.AdminNotes = "approved", adminNotes
	if err := s.db.Save(application).Error; err != nil {
		return nil, err
	}
	html := renderApprovalEmail(
		application.ApplicantName, application.Server.Hostname, result["ssh_port"].(int),
		result["extra_ports"].([]int), result["password"].(string), adminNotes,
	)
	s.notify(application.ApplicantEmail, html)
	return applicationResponse(application), nil
}

func (s *ApplicationService) Reject(id uint, adminNotes string) (*dto.ApplicationResponse, error) {
	application, err := s.find(id)
	if err != nil {
		return nil, ErrApplicationNotFound
	}
	if application.Status != "pending" {
		return nil, ErrApplicationNotPending
	}
	application.Status, application.AdminNotes = "rejected", adminNotes
	if err := s.db.Save(application).Error; err != nil {
		return nil, err
	}
	s.notify(application.ApplicantEmail,
		renderRejectionEmail(application.ApplicantName, application.Server.Host, application.Image.Name, adminNotes))
	return applicationResponse(application), nil
}

func (s *ApplicationService) ListPublicServers() ([]dto.PublicServerInfo, error) {
	servers, err := s.serverService.List()
	if err != nil {
		return nil, err
	}
	result := make([]dto.PublicServerInfo, len(servers))
	for i, server := range servers {
		result[i] = dto.PublicServerInfo{ID: server.ID, Host: server.Host, Description: server.Description}
	}
	return result, nil
}

func (s *ApplicationService) ListPublicImages(serverID uint) ([]dto.PublicImageInfo, error) {
	var images []model.Image
	if err := s.db.Where("server_id = ?", serverID).Order("id desc").Find(&images).Error; err != nil {
		return nil, err
	}
	result := make([]dto.PublicImageInfo, len(images))
	for i, image := range images {
		result[i] = dto.PublicImageInfo{ID: image.ID, Name: image.Name, ImageAddress: image.ImageAddress}
	}
	return result, nil
}

func (s *ApplicationService) find(id uint) (*model.Application, error) {
	var application model.Application
	if err := s.db.Preload("Server").First(&application, id).Error; err != nil {
		return nil, err
	}
	return &application, s.db.First(&application.Image, application.ImageID).Error
}

func (s *ApplicationService) notify(to, html string) {
	if strings.TrimSpace(to) != "" && s.sendAsync != nil {
		s.sendAsync(to, "[Server Dock]", html)
	}
}

func applicationResponse(application *model.Application) *dto.ApplicationResponse {
	return &dto.ApplicationResponse{
		ID: application.ID, ApplicantName: application.ApplicantName, ApplicantEmail: application.ApplicantEmail,
		ServerID: application.ServerID, ServerHost: application.Server.Host,
		ImageID: application.ImageID, ImageName: application.Image.Name,
		Status: application.Status, AdminNotes: application.AdminNotes,
		CreatedAt: application.CreatedAt, UpdatedAt: application.UpdatedAt,
	}
}
