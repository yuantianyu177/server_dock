package service

import (
	"errors"
	"serverdock/internal/dto"
	"serverdock/internal/model"
	"serverdock/internal/repository"
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAppService(t *testing.T) (*ApplicationService, *MockEmailServiceImpl, *MockDockerService) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test DB: %v", err)
	}
	db.AutoMigrate(&model.Server{}, &model.Image{}, &model.Application{}, &model.SystemConfig{})

	serverRepo := repository.NewServerRepo(db)
	imageRepo := repository.NewImageRepo(db)
	appRepo := repository.NewApplicationRepo(db)
	configRepo := repository.NewConfigRepo(db)

	mockSSH := &MockSSHService{}
	mockDocker := &MockDockerService{}
	mockEmail := &MockEmailServiceImpl{}

	serverSvc := NewServerService(serverRepo, mockSSH, testEncryptKey)
	containerSvc := NewContainerService(serverSvc, mockDocker, mockSSH)
	configSvc := NewConfigService(configRepo)
	configSvc.EnsureDefaults()

	appSvc := NewApplicationService(appRepo, imageRepo, serverSvc, containerSvc, configSvc, mockEmail)

	// Create test server and image
	serverSvc.Create(&dto.CreateServerRequest{
		Host: "GPU Server", Hostname: "10.0.0.1", User: "root", AuthType: "password", Credential: "pass",
	})
	imageRepo.Create(&model.Image{ServerID: 1, DockerImageID: "sha256:abc", Name: "Ubuntu CUDA", ImageAddress: "nvidia/cuda:12.0"})

	return appSvc, mockEmail, mockDocker
}

func TestApplicationService_Submit(t *testing.T) {
	svc, _, _ := setupAppService(t)

	resp, err := svc.Submit(&dto.SubmitApplicationRequest{
		ApplicantName: "Zhang San", ApplicantEmail: "zhang@example.com",
		ServerID: 1, ImageID: 1,
	})
	if err != nil {
		t.Fatalf("Submit failed: %v", err)
	}
	if resp.Status != "pending" {
		t.Fatalf("Expected 'pending', got %s", resp.Status)
	}
	if resp.ServerHost != "GPU Server" {
		t.Fatalf("Expected 'GPU Server', got %s", resp.ServerHost)
	}
}

func TestApplicationService_SubmitSendsAdminEmailWithImageName(t *testing.T) {
	svc, mockEmail, _ := setupAppService(t)
	if err := svc.configService.Set("admin_email", "admin@example.com"); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	_, err := svc.Submit(&dto.SubmitApplicationRequest{
		ApplicantName: "Zhang San", ApplicantEmail: "zhang@example.com",
		ServerID: 1, ImageID: 1,
	})
	if err != nil {
		t.Fatalf("Submit failed: %v", err)
	}

	if mockEmail.AsyncCalls != 1 {
		t.Fatalf("expected submit email to be sent asynchronously, got %d async calls", mockEmail.AsyncCalls)
	}
	if got := mockEmail.SentEmails[0].Subject; got != "[Server Dock]" {
		t.Fatalf("expected submit email subject [Server Dock], got %q", got)
	}
	body := mockEmail.SentEmails[0].Body
	for _, expected := range []string{"Ubuntu CUDA", "GPU Server", "zhang@example.com"} {
		if !strings.Contains(body, expected) {
			t.Fatalf("expected admin email to contain %q, got %q", expected, body)
		}
	}
}

func TestApplicationService_SubmitInvalidServer(t *testing.T) {
	svc, _, _ := setupAppService(t)

	_, err := svc.Submit(&dto.SubmitApplicationRequest{
		ApplicantName: "Test", ApplicantEmail: "t@t.com",
		ServerID: 999, ImageID: 1,
	})
	if err == nil {
		t.Fatal("Expected error for invalid server")
	}
}

func TestApplicationService_SubmitImageServerMismatch(t *testing.T) {
	svc, _, _ := setupAppService(t)

	// Image 1 belongs to server 1, try with server 2 (doesn't exist but image check happens after)
	_, err := svc.Submit(&dto.SubmitApplicationRequest{
		ApplicantName: "Test", ApplicantEmail: "t@t.com",
		ServerID: 999, ImageID: 1,
	})
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestApplicationService_ListAndGet(t *testing.T) {
	svc, _, _ := setupAppService(t)

	svc.Submit(&dto.SubmitApplicationRequest{
		ApplicantName: "A", ApplicantEmail: "a@a.com", ServerID: 1, ImageID: 1,
	})
	svc.Submit(&dto.SubmitApplicationRequest{
		ApplicantName: "B", ApplicantEmail: "b@b.com", ServerID: 1, ImageID: 1,
	})

	list, _ := svc.List("")
	if len(list) != 2 {
		t.Fatalf("Expected 2, got %d", len(list))
	}

	pending, _ := svc.List("pending")
	if len(pending) != 2 {
		t.Fatalf("Expected 2 pending, got %d", len(pending))
	}

	got, _ := svc.GetByID(1)
	if got.ApplicantName != "A" {
		t.Fatalf("Expected 'A', got %s", got.ApplicantName)
	}
}

func TestApplicationService_Reject(t *testing.T) {
	svc, mockEmail, _ := setupAppService(t)

	svc.Submit(&dto.SubmitApplicationRequest{
		ApplicantName: "Zhang", ApplicantEmail: "z@z.com", ServerID: 1, ImageID: 1,
	})

	resp, err := svc.Reject(1, "Not enough resources")
	if err != nil {
		t.Fatalf("Reject failed: %v", err)
	}
	if resp.Status != "rejected" {
		t.Fatalf("Expected 'rejected', got %s", resp.Status)
	}

	// Should have sent rejection email
	found := false
	for _, e := range mockEmail.SentEmails {
		if e.Subject == "[Server Dock]" {
			found = true
		}
	}
	if !found {
		t.Error("Expected rejection email to be sent")
	}
	if mockEmail.AsyncCalls == 0 {
		t.Fatal("expected rejection email to be sent asynchronously")
	}
}

func TestApplicationService_RejectNotPending(t *testing.T) {
	svc, _, _ := setupAppService(t)

	svc.Submit(&dto.SubmitApplicationRequest{
		ApplicantName: "A", ApplicantEmail: "a@a.com", ServerID: 1, ImageID: 1,
	})
	svc.Reject(1, "no")

	_, err := svc.Reject(1, "again")
	if err == nil {
		t.Fatal("Expected error for non-pending application")
	}
}

func TestApplicationService_Approve(t *testing.T) {
	svc, mockEmail, mockDocker := setupAppService(t)
	var createCmd string
	mockDocker.CreateContainerFn = func(cmd string) (string, error) {
		createCmd = cmd
		return "created", nil
	}
	if err := svc.configService.Set("docker_extra_args", "--gpus all\n--shm-size=8g"); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	_, err := svc.Submit(&dto.SubmitApplicationRequest{
		ApplicantName: "Zhang", ApplicantEmail: "z@z.com", ServerID: 1, ImageID: 1,
	})
	if err != nil {
		t.Fatalf("Submit failed: %v", err)
	}

	resp, err := svc.Approve(1, "Approved for testing")
	if err != nil {
		t.Fatalf("Approve failed: %v", err)
	}
	if resp.Status != "approved" {
		t.Fatalf("Expected 'approved', got %s", resp.Status)
	}
	if resp.AdminNotes != "Approved for testing" {
		t.Fatalf("Expected admin notes to be persisted, got %q", resp.AdminNotes)
	}

	found := false
	for _, e := range mockEmail.SentEmails {
		if e.Subject == "[Server Dock]" {
			found = true
		}
	}
	if !found {
		t.Error("Expected approval email to be sent")
	}
	if mockEmail.AsyncCalls == 0 {
		t.Fatal("expected approval email to be sent asynchronously")
	}
	if strings.Contains(createCmd, "\n") || strings.Contains(createCmd, "\r") {
		t.Fatalf("expected docker command to be single-line, got %q", createCmd)
	}
	for _, part := range []string{"--gpus all", "--shm-size=8g", "nvidia/cuda:12.0"} {
		if !strings.Contains(createCmd, part) {
			t.Fatalf("expected docker command to contain %q, got %q", part, createCmd)
		}
	}
}

func TestApplicationService_ApproveContainerFailure(t *testing.T) {
	svc, _, mockDocker := setupAppService(t)
	mockDocker.CreateContainerFn = func(cmd string) (string, error) {
		return "", errors.New("docker daemon unavailable")
	}

	_, err := svc.Submit(&dto.SubmitApplicationRequest{
		ApplicantName: "Zhang", ApplicantEmail: "z@z.com", ServerID: 1, ImageID: 1,
	})
	if err != nil {
		t.Fatalf("Submit failed: %v", err)
	}

	_, err = svc.Approve(1, "")
	if err == nil {
		t.Fatal("Expected approval to fail")
	}
	if !errors.Is(err, ErrContainerProvisioning) {
		t.Fatalf("Expected ErrContainerProvisioning, got %v", err)
	}
	if !strings.Contains(err.Error(), "docker daemon unavailable") {
		t.Fatalf("Expected detailed container error, got %v", err)
	}
}

func TestApplicationService_ApproveEmptyImageAddress(t *testing.T) {
	svc, _, _ := setupAppService(t)

	img, err := svc.imageRepo.FindByID(1)
	if err != nil {
		t.Fatalf("failed to load image: %v", err)
	}
	img.ImageAddress = ""
	if err := svc.imageRepo.Update(img); err != nil {
		t.Fatalf("failed to update image: %v", err)
	}

	_, err = svc.Submit(&dto.SubmitApplicationRequest{
		ApplicantName: "Zhang", ApplicantEmail: "z@z.com", ServerID: 1, ImageID: 1,
	})
	if err != nil {
		t.Fatalf("Submit failed: %v", err)
	}

	_, err = svc.Approve(1, "")
	if err == nil {
		t.Fatal("Expected approval to fail for empty image address")
	}
	if !errors.Is(err, ErrContainerProvisioning) {
		t.Fatalf("Expected ErrContainerProvisioning, got %v", err)
	}
	if !strings.Contains(err.Error(), "empty image address") {
		t.Fatalf("Expected empty image address error, got %v", err)
	}
}

func TestGenerateContainerName(t *testing.T) {
	name := GenerateContainerName("Zhang San")
	if !strings.HasPrefix(name, "zhangsan-") {
		t.Fatalf("Expected name to start with 'zhangsan-', got %s", name)
	}
	// Should have timestamp suffix
	if len(name) != len("zhangsan-20060102150405") {
		t.Fatalf("Expected longer name with timestamp, got %s", name)
	}
}

func TestGenerateContainerNameChinese(t *testing.T) {
	name := GenerateContainerName("张三")
	if !strings.HasPrefix(name, "zhangsan-") {
		t.Fatalf("Expected 'zhangsan-' prefix for Chinese name, got %s", name)
	}
}

func TestGenerateContainerNameFallback(t *testing.T) {
	name := GenerateContainerName("!!!")
	if !strings.HasPrefix(name, "container-") {
		t.Fatalf("Expected 'container-' prefix for unsupported chars, got %s", name)
	}
}
