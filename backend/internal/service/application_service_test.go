package service

import (
	"errors"
	"strings"
	"testing"

	"serverdock/internal/dto"
	"serverdock/internal/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAppService(t *testing.T) (*ApplicationService, *MockEmailServiceImpl, *MockSSHService) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test DB: %v", err)
	}
	if err := db.AutoMigrate(&model.Server{}, &model.Image{}, &model.Application{}, &model.SystemConfig{}); err != nil {
		t.Fatalf("migrate test DB: %v", err)
	}

	mockSSH := &MockSSHService{}
	mockEmail := &MockEmailServiceImpl{}
	servers := NewServerService(db, mockSSH.TestConnection, mockSSH.ExecuteCommand, testEncryptKey)
	containers := NewContainerService(servers)
	config := NewConfigService(db)
	if err := config.EnsureDefaults(); err != nil {
		t.Fatalf("create default config: %v", err)
	}
	service := NewApplicationService(db, servers, containers, config, mockEmail.SendAsync)

	if _, err := servers.Create(&dto.CreateServerRequest{
		Host: "GPU Server", Hostname: "10.0.0.1", User: "root", AuthType: "password", Credential: "pass",
	}); err != nil {
		t.Fatalf("create test server: %v", err)
	}
	if err := db.Create(&model.Image{
		ServerID: 1, DockerImageID: "sha256:abc", Name: "Ubuntu CUDA", ImageAddress: "nvidia/cuda:12.0",
	}).Error; err != nil {
		t.Fatalf("create test image: %v", err)
	}
	return service, mockEmail, mockSSH
}

func submitTestApplication(t *testing.T, service *ApplicationService) *dto.ApplicationResponse {
	t.Helper()
	response, err := service.Submit(&dto.SubmitApplicationRequest{
		ApplicantName: "Zhang San", ApplicantEmail: "zhang@example.com", ServerID: 1, ImageID: 1,
	})
	if err != nil {
		t.Fatalf("submit application: %v", err)
	}
	return response
}

func TestApplicationServiceSubmit(t *testing.T) {
	service, _, _ := setupAppService(t)
	response := submitTestApplication(t, service)
	if response.Status != "pending" || response.ServerHost != "GPU Server" {
		t.Fatalf("unexpected response: %+v", response)
	}
}

func TestApplicationServiceSubmitNotifiesAdmin(t *testing.T) {
	service, email, _ := setupAppService(t)
	if err := service.configService.Set("admin_email", "admin@example.com"); err != nil {
		t.Fatal(err)
	}
	submitTestApplication(t, service)

	if email.AsyncCalls != 1 {
		t.Fatalf("expected one email, got %d", email.AsyncCalls)
	}
	for _, value := range []string{"Ubuntu CUDA", "GPU Server", "zhang@example.com"} {
		if !strings.Contains(email.SentEmails[0].Body, value) {
			t.Fatalf("email does not contain %q", value)
		}
	}
}

func TestApplicationServiceSubmitRejectsInvalidSelection(t *testing.T) {
	service, _, _ := setupAppService(t)
	if _, err := service.Submit(&dto.SubmitApplicationRequest{
		ApplicantName: "Test", ApplicantEmail: "test@example.com", ServerID: 999, ImageID: 1,
	}); err == nil {
		t.Fatal("expected invalid server error")
	}

	server, err := service.serverService.Create(&dto.CreateServerRequest{
		Host: "Other", Hostname: "10.0.0.2", User: "root", AuthType: "password", Credential: "pass",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Submit(&dto.SubmitApplicationRequest{
		ApplicantName: "Test", ApplicantEmail: "test@example.com", ServerID: server.ID, ImageID: 1,
	}); err == nil {
		t.Fatal("expected image/server mismatch error")
	}
}

func TestApplicationServiceList(t *testing.T) {
	service, _, _ := setupAppService(t)
	submitTestApplication(t, service)
	if _, err := service.Submit(&dto.SubmitApplicationRequest{
		ApplicantName: "Li Si", ApplicantEmail: "li@example.com", ServerID: 1, ImageID: 1,
	}); err != nil {
		t.Fatal(err)
	}
	applications, err := service.List()
	if err != nil || len(applications) != 2 {
		t.Fatalf("expected two applications, got %d (%v)", len(applications), err)
	}
	if applications[0].ServerHost != "GPU Server" || applications[0].ImageName != "Ubuntu CUDA" {
		t.Fatalf("missing joined details: %+v", applications[0])
	}
}

func TestApplicationServiceReject(t *testing.T) {
	service, email, _ := setupAppService(t)
	application := submitTestApplication(t, service)
	response, err := service.Reject(application.ID, "Not enough resources")
	if err != nil {
		t.Fatal(err)
	}
	if response.Status != "rejected" || email.AsyncCalls != 1 {
		t.Fatalf("unexpected rejection result: %+v, emails=%d", response, email.AsyncCalls)
	}
	if _, err := service.Reject(application.ID, "again"); !errors.Is(err, ErrApplicationNotPending) {
		t.Fatalf("expected ErrApplicationNotPending, got %v", err)
	}
}

func TestApplicationServiceApprove(t *testing.T) {
	service, email, ssh := setupAppService(t)
	var runCommand string
	ssh.ExecuteCommandFn = func(_ string, _ int, _, _, _, command string) (string, error) {
		if strings.HasPrefix(command, "docker run ") {
			runCommand = command
		}
		return "created", nil
	}
	if err := service.configService.Set("docker_extra_args", "--gpus all\n--shm-size=8g"); err != nil {
		t.Fatal(err)
	}
	application := submitTestApplication(t, service)

	response, err := service.Approve(application.ID, "Approved for testing")
	if err != nil {
		t.Fatal(err)
	}
	if response.Status != "approved" || response.AdminNotes != "Approved for testing" || email.AsyncCalls != 1 {
		t.Fatalf("unexpected approval result: %+v, emails=%d", response, email.AsyncCalls)
	}
	if strings.ContainsAny(runCommand, "\r\n") {
		t.Fatalf("docker command should be one line: %q", runCommand)
	}
	for _, value := range []string{"--gpus all", "--shm-size=8g", "nvidia/cuda:12.0"} {
		if !strings.Contains(runCommand, value) {
			t.Fatalf("docker command does not contain %q: %s", value, runCommand)
		}
	}
}

func TestApplicationServiceApproveContainerFailure(t *testing.T) {
	service, _, ssh := setupAppService(t)
	ssh.ExecuteCommandFn = func(_ string, _ int, _, _, _, command string) (string, error) {
		if strings.HasPrefix(command, "docker run ") {
			return "", errors.New("docker daemon unavailable")
		}
		return "", nil
	}
	application := submitTestApplication(t, service)
	_, err := service.Approve(application.ID, "")
	if !errors.Is(err, ErrContainerProvisioning) || !strings.Contains(err.Error(), "docker daemon unavailable") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApplicationServiceApproveRejectsEmptyImageAddress(t *testing.T) {
	service, _, _ := setupAppService(t)
	if err := service.db.Model(&model.Image{}).Where("id = ?", 1).Update("image_address", "").Error; err != nil {
		t.Fatal(err)
	}
	application := submitTestApplication(t, service)
	_, err := service.Approve(application.ID, "")
	if !errors.Is(err, ErrContainerProvisioning) || !strings.Contains(err.Error(), "empty image address") {
		t.Fatalf("unexpected error: %v", err)
	}
}
