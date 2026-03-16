package service

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"serverdock/internal/dto"
	"serverdock/internal/model"
	"serverdock/internal/repository"

	"github.com/mozillazg/go-pinyin"
)

var (
	ErrApplicationNotFound   = errors.New("application not found")
	ErrApplicationNotPending = errors.New("application is not pending")
	ErrContainerProvisioning = errors.New("failed to provision container")
)

type ApplicationService struct {
	appRepo          *repository.ApplicationRepo
	imageRepo        *repository.ImageRepo
	serverService    *ServerService
	containerService *ContainerService
	configService    *ConfigService
	emailService     EmailService
}

func NewApplicationService(
	appRepo *repository.ApplicationRepo,
	imageRepo *repository.ImageRepo,
	serverService *ServerService,
	containerService *ContainerService,
	configService *ConfigService,
	emailService EmailService,
) *ApplicationService {
	return &ApplicationService{
		appRepo:          appRepo,
		imageRepo:        imageRepo,
		serverService:    serverService,
		containerService: containerService,
		configService:    configService,
		emailService:     emailService,
	}
}

// Submit creates a new application with pending status.
func (s *ApplicationService) Submit(req *dto.SubmitApplicationRequest) (*dto.ApplicationResponse, error) {
	// Verify server exists
	server, err := s.serverService.GetByID(req.ServerID)
	if err != nil {
		return nil, errors.New("server not found")
	}

	img, err := s.imageRepo.FindByID(req.ImageID)
	if err != nil {
		return nil, errors.New("image not found")
	}
	if img.ServerID != req.ServerID {
		return nil, errors.New("image does not belong to the selected server")
	}

	app := &model.Application{
		ApplicantName:  req.ApplicantName,
		ApplicantEmail: req.ApplicantEmail,
		ServerID:       req.ServerID,
		ImageID:        req.ImageID,
		Status:         "pending",
	}

	if err := s.appRepo.Create(app); err != nil {
		return nil, err
	}

	// Load relations
	app, _ = s.appRepo.FindByID(app.ID)

	// Notify admin
	adminEmail := s.configService.Get("admin_email")
	if adminEmail != "" {
		serverHost := ""
		if server != nil {
			serverHost = server.Host
		}
		html := RenderNewApplicationEmail(app.ApplicantName, app.ApplicantEmail, serverHost, img.Name)
		s.sendNotification(adminEmail, "[Server Dock]", html)
	}

	return s.toResponse(app), nil
}

// List returns applications filtered by status.
func (s *ApplicationService) List(status string) ([]dto.ApplicationResponse, error) {
	apps, err := s.appRepo.List(status)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.ApplicationResponse, 0, len(apps))
	for _, app := range apps {
		responses = append(responses, *s.toResponse(&app))
	}
	return responses, nil
}

// GetByID returns an application by ID.
func (s *ApplicationService) GetByID(id uint) (*dto.ApplicationResponse, error) {
	app, err := s.appRepo.FindByID(id)
	if err != nil {
		return nil, ErrApplicationNotFound
	}
	return s.toResponse(app), nil
}

// Approve approves an application and creates a container.
func (s *ApplicationService) Approve(id uint, adminNotes string) (*dto.ApplicationResponse, error) {
	app, err := s.appRepo.FindByID(id)
	if err != nil {
		return nil, ErrApplicationNotFound
	}
	if app.Status != "pending" {
		return nil, ErrApplicationNotPending
	}

	// Generate container name: applicant name + timestamp
	containerName := GenerateContainerName(app.ApplicantName)

	// Use preloaded Image from FindByID; fall back to direct query if not populated
	imageAddress := strings.TrimSpace(app.Image.ImageAddress)
	if imageAddress == "" {
		img, err := s.imageRepo.FindByID(app.ImageID)
		if err != nil {
			return nil, fmt.Errorf("%w: image %d not found", ErrContainerProvisioning, app.ImageID)
		}
		imageAddress = strings.TrimSpace(img.ImageAddress)
	}
	if imageAddress == "" {
		return nil, fmt.Errorf("%w: image %d has empty image address", ErrContainerProvisioning, app.ImageID)
	}

	// Get all config values in a single DB query
	cfgMap, err := s.configService.GetAllAsMap()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	portStart, _ := strconv.Atoi(cfgMap["port_range_start"])
	portEnd, _ := strconv.Atoi(cfgMap["port_range_end"])
	extraPorts, _ := strconv.Atoi(cfgMap["extra_ports_per_container"])
	mountPath := cfgMap["default_volume_mount_path"]
	dockerExtraArgs := cfgMap["docker_extra_args"]

	// Create container
	result, err := s.containerService.CreateContainer(
		app.ServerID, containerName, imageAddress, dockerExtraArgs,
		portStart, portEnd, extraPorts, mountPath,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrContainerProvisioning, err)
	}

	// Update application status
	app.Status = "approved"
	app.AdminNotes = adminNotes
	if err := s.appRepo.Update(app); err != nil {
		return nil, err
	}

	// Send approval email
	sshPort := result["ssh_port"].(int)
	extraPortsList := result["extra_ports"].([]int)
	password := result["password"].(string)
	serverHostname := app.Server.Hostname

	html := RenderApprovalEmail(app.ApplicantName, serverHostname, sshPort, extraPortsList, password, adminNotes)
	s.sendNotification(app.ApplicantEmail, "[Server Dock]", html)

	return s.toResponse(app), nil
}

// Reject rejects an application.
func (s *ApplicationService) Reject(id uint, adminNotes string) (*dto.ApplicationResponse, error) {
	app, err := s.appRepo.FindByID(id)
	if err != nil {
		return nil, ErrApplicationNotFound
	}
	if app.Status != "pending" {
		return nil, ErrApplicationNotPending
	}

	app.Status = "rejected"
	app.AdminNotes = adminNotes
	if err := s.appRepo.Update(app); err != nil {
		return nil, err
	}

	// Send rejection email (Server and Image are preloaded by FindByID)
	html := RenderRejectionEmail(app.ApplicantName, app.Server.Host, app.Image.Name, adminNotes)
	s.sendNotification(app.ApplicantEmail, "[Server Dock]", html)

	return s.toResponse(app), nil
}

// ListPublicServers returns servers with container count for public view.
func (s *ApplicationService) ListPublicServers() ([]dto.PublicServerInfo, error) {
	servers, err := s.serverService.List()
	if err != nil {
		return nil, err
	}

	result := make([]dto.PublicServerInfo, 0, len(servers))

	// Fetch container counts concurrently with a timeout
	type countResult struct {
		idx   int
		count int
	}
	ch := make(chan countResult, len(servers))

	for i, srv := range servers {
		go func(idx int, id uint) {
			count := 0
			containers, err := s.containerService.ListContainers(id)
			if err == nil {
				count = len(containers)
			}
			ch <- countResult{idx: idx, count: count}
		}(i, srv.ID)
	}

	counts := make(map[int]int, len(servers))
	// Wait with a 5-second timeout to avoid blocking the frontend
	timeout := time.After(5 * time.Second)
	remaining := len(servers)
	for remaining > 0 {
		select {
		case r := <-ch:
			counts[r.idx] = r.count
			remaining--
		case <-timeout:
			remaining = 0
		}
	}

	for i, srv := range servers {
		result = append(result, dto.PublicServerInfo{
			ID:             srv.ID,
			Host:           srv.Host,
			Description:    srv.Description,
			ContainerCount: counts[i],
		})
	}
	return result, nil
}

// ListPublicImages returns available images for a server.
func (s *ApplicationService) ListPublicImages(serverID uint) ([]dto.PublicImageInfo, error) {
	images, err := s.imageRepo.List(&serverID)
	if err != nil {
		return nil, err
	}

	var result []dto.PublicImageInfo
	for _, img := range images {
		result = append(result, dto.PublicImageInfo{
			ID:           img.ID,
			Name:         img.Name,
			ImageAddress: img.ImageAddress,
		})
	}
	return result, nil
}

// GenerateContainerName generates a container name from applicant name + timestamp.
func GenerateContainerName(applicantName string) string {
	args := pinyin.NewArgs()
	args.Style = pinyin.Normal
	args.Separator = ""
	args.Fallback = func(r rune, _ pinyin.Args) []string {
		switch {
		case r >= 'a' && r <= 'z':
			return []string{string(r)}
		case r >= 'A' && r <= 'Z':
			return []string{strings.ToLower(string(r))}
		case r >= '0' && r <= '9':
			return []string{string(r)}
		default:
			return []string{}
		}
	}

	name := strings.Join(pinyin.LazyPinyin(applicantName, args), "")
	if name == "" {
		name = "container"
	}
	timestamp := time.Now().Format("20060102150405")
	return fmt.Sprintf("%s-%s", name, timestamp)
}

func (s *ApplicationService) sendNotification(to, subject, html string) {
	if strings.TrimSpace(to) == "" {
		return
	}
	s.emailService.SendAsync(to, subject, html)
}

func (s *ApplicationService) toResponse(app *model.Application) *dto.ApplicationResponse {
	resp := &dto.ApplicationResponse{
		ID:             app.ID,
		ApplicantName:  app.ApplicantName,
		ApplicantEmail: app.ApplicantEmail,
		ServerID:       app.ServerID,
		ImageID:        app.ImageID,
		Status:         app.Status,
		AdminNotes:     app.AdminNotes,
		CreatedAt:      app.CreatedAt,
		UpdatedAt:      app.UpdatedAt,
	}
	if app.Server.ID != 0 {
		resp.ServerHost = app.Server.Host
	} else if app.ServerID != 0 {
		if server, err := s.serverService.GetByID(app.ServerID); err == nil {
			resp.ServerHost = server.Host
		}
	}
	if app.Image.ID != 0 {
		resp.ImageName = app.Image.Name
	} else if app.ImageID != 0 {
		if image, err := s.imageRepo.FindByID(app.ImageID); err == nil {
			resp.ImageName = image.Name
		}
	}
	return resp
}
