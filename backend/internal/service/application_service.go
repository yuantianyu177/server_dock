package service

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"serverdock/internal/dto"
	"serverdock/internal/model"

	"gorm.io/gorm"
)

var (
	ErrApplicationNotFound   = errors.New("application not found")
	ErrApplicationNotPending = errors.New("application is not pending")
	ErrContainerProvisioning = errors.New("container provisioning failed")
	ErrInvalidEmailAction    = errors.New("invalid or expired email action token")
)

type ApplicationService struct {
	db                *gorm.DB
	serverService     *ServerService
	containerService  *ContainerService
	configService     *ConfigService
	emailActionSecret string
	fallbackPublicURL string
	now               func() time.Time
	sendAsync         func(to, subject, html string)
}

func NewApplicationService(db *gorm.DB, serverService *ServerService, containerService *ContainerService, configService *ConfigService, emailActionSecret, fallbackPublicURL string, sendAsync func(string, string, string)) *ApplicationService {
	return &ApplicationService{
		db: db, serverService: serverService, containerService: containerService,
		configService: configService, emailActionSecret: emailActionSecret,
		fallbackPublicURL: fallbackPublicURL, now: time.Now, sendAsync: sendAsync,
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
		actions := s.newEmailActionLinks(application.ID)
		s.notify(to, renderNewApplicationEmail(application.ApplicantName, application.ApplicantEmail, server.Host, image.Name, actions))
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

func (s *ApplicationService) Approve(id uint) (*dto.ApplicationResponse, error) {
	application, err := s.findPending(id)
	if err != nil {
		return nil, err
	}
	if application.Image.ID == 0 {
		return nil, fmt.Errorf("%w: image %d not found", ErrContainerProvisioning, application.ImageID)
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
	extraPortCount, _ := strconv.Atoi(config["extra_ports_per_container"])
	result, err := s.containerService.CreateContainer(
		application.ServerID, fmt.Sprintf("container-%d", application.ID), imageAddress,
		config["docker_extra_args"], portStart, portEnd, extraPortCount, config["default_volume_mount_path"],
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrContainerProvisioning, err)
	}

	if err := s.updateStatus(application, "approved"); err != nil {
		return nil, err
	}
	server := application.Server.Hostname
	sshPort := result["ssh_port"].(int)
	extraPorts := result["extra_ports"].([]int)
	password := result["password"].(string)
	html := renderApprovalEmail(application.ApplicantName, server, sshPort, extraPorts, password)
	s.notify(application.ApplicantEmail, html)
	response := applicationResponse(application)
	response.ConnectionInfo = &dto.ContainerConnectionInfo{
		Server:     server,
		User:       "root",
		Password:   password,
		SSHPort:    sshPort,
		ExtraPorts: formatPorts(extraPorts),
		SSHCommand: fmt.Sprintf("ssh -p %d root@%s", sshPort, server),
	}
	return response, nil
}

func (s *ApplicationService) Reject(id uint) (*dto.ApplicationResponse, error) {
	application, err := s.findPending(id)
	if err != nil {
		return nil, err
	}
	if err := s.updateStatus(application, "rejected"); err != nil {
		return nil, err
	}
	serverName := application.Server.Host
	if serverName == "" {
		serverName = fmt.Sprintf("server #%d", application.ServerID)
	}
	imageName := application.Image.Name
	if imageName == "" {
		imageName = fmt.Sprintf("image #%d", application.ImageID)
	}
	s.notify(application.ApplicantEmail,
		renderRejectionEmail(application.ApplicantName, serverName, imageName))
	return applicationResponse(application), nil
}

func (s *ApplicationService) Ignore(id uint) (*dto.ApplicationResponse, error) {
	application, err := s.findPending(id)
	if err != nil {
		return nil, err
	}
	if err := s.updateStatus(application, "ignored"); err != nil {
		return nil, err
	}
	return applicationResponse(application), nil
}

func (s *ApplicationService) ListPublicServers() ([]dto.PublicServerInfo, error) {
	servers, err := s.serverService.List()
	if err != nil {
		return nil, err
	}
	result := make([]dto.PublicServerInfo, len(servers))
	var waitGroup sync.WaitGroup
	concurrency := make(chan struct{}, 4)

	for i, server := range servers {
		result[i] = dto.PublicServerInfo{ID: server.ID, Host: server.Host, Description: server.Description}
		waitGroup.Add(1)
		go func(index int, serverID uint) {
			defer waitGroup.Done()
			concurrency <- struct{}{}
			defer func() { <-concurrency }()

			containers, listErr := s.containerService.ListContainers(serverID)
			if listErr != nil {
				return
			}

			result[index].LoadAvailable = true
			result[index].TotalContainers = len(containers)
			for _, container := range containers {
				status := strings.ToLower(strings.TrimSpace(container["status"]))
				if strings.HasPrefix(status, "up") {
					result[index].RunningContainers++
				}
			}
		}(i, server.ID)
	}

	waitGroup.Wait()
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
	if err := s.db.First(&application.Image, application.ImageID).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return &application, nil
}

func (s *ApplicationService) findPending(id uint) (*model.Application, error) {
	application, err := s.find(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrApplicationNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to load application: %w", err)
	}
	if application.Status != "pending" {
		return nil, ErrApplicationNotPending
	}
	return application, nil
}

func (s *ApplicationService) updateStatus(application *model.Application, status string) error {
	result := s.db.Model(application).Where("status = ?", "pending").Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrApplicationNotPending
	}
	application.Status = status
	return nil
}

func (s *ApplicationService) notify(to, html string) {
	if strings.TrimSpace(to) != "" && s.sendAsync != nil {
		s.sendAsync(to, notificationSubject, html)
	}
}

func applicationResponse(application *model.Application) *dto.ApplicationResponse {
	return &dto.ApplicationResponse{
		ID: application.ID, ApplicantName: application.ApplicantName, ApplicantEmail: application.ApplicantEmail,
		ServerID: application.ServerID, ServerHost: application.Server.Host,
		ImageID: application.ImageID, ImageName: application.Image.Name,
		Status:    application.Status,
		CreatedAt: application.CreatedAt, UpdatedAt: application.UpdatedAt,
	}
}
