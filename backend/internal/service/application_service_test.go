package service

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

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
	service := NewApplicationService(
		db,
		servers,
		containers,
		config,
		"test-email-action-secret",
		"http://serverdock.test",
		mockEmail.SendAsync,
	)

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
	for _, value := range []string{
		"Ubuntu CUDA", "GPU Server", "zhang@example.com",
		"忽略", "拒绝", "批准",
		"http://serverdock.test/api/applications/public/email-action?action=ignore#token=",
		"http://serverdock.test/api/applications/public/email-action?action=reject#token=",
		"http://serverdock.test/api/applications/public/email-action?action=approve#token=",
	} {
		if !strings.Contains(email.SentEmails[0].Body, value) {
			t.Fatalf("email does not contain %q", value)
		}
	}
}

func TestApplicationServiceEmailActions(t *testing.T) {
	tests := []struct {
		action string
		status string
	}{
		{action: "ignore", status: "ignored"},
		{action: "reject", status: "rejected"},
		{action: "approve", status: "approved"},
	}

	for _, test := range tests {
		t.Run(test.action, func(t *testing.T) {
			service, _, _ := setupAppService(t)
			application := submitTestApplication(t, service)
			token, err := service.createEmailActionToken(application.ID, test.action)
			if err != nil {
				t.Fatalf("create email action token: %v", err)
			}

			response, err := service.HandleEmailAction(token)
			if err != nil {
				t.Fatalf("handle email action: %v", err)
			}
			if response.Status != test.status {
				t.Fatalf("expected status %q, got %q", test.status, response.Status)
			}
			adminApplications, err := service.List()
			if err != nil {
				t.Fatalf("load admin applications after email action: %v", err)
			}
			if len(adminApplications) != 1 || adminApplications[0].Status != test.status {
				t.Fatalf("admin list did not reflect email action: %+v", adminApplications)
			}
			if _, err := service.HandleEmailAction(token); !errors.Is(err, ErrApplicationNotPending) {
				t.Fatalf("expected one-time link to be rejected, got %v", err)
			}
		})
	}
}

func TestApplicationServiceEmailActionRejectsInvalidAndExpiredTokens(t *testing.T) {
	service, _, _ := setupAppService(t)
	application := submitTestApplication(t, service)
	now := time.Date(2026, time.August, 27, 8, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }
	token, err := service.createEmailActionToken(application.ID, "ignore")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := service.HandleEmailAction(token + "tampered"); !errors.Is(err, ErrInvalidEmailAction) {
		t.Fatalf("expected tampered token to fail, got %v", err)
	}
	service.now = func() time.Time { return now.Add(emailActionTokenTTL + time.Second) }
	if _, err := service.HandleEmailAction(token); !errors.Is(err, ErrInvalidEmailAction) {
		t.Fatalf("expected expired token to fail, got %v", err)
	}
}

func TestApplicationServiceEmailActionsUseConfiguredPublicURL(t *testing.T) {
	service, email, _ := setupAppService(t)
	if err := service.configService.Set("admin_email", "admin@example.com"); err != nil {
		t.Fatal(err)
	}
	if err := service.configService.Set("public_url", "https://dock.example.com/base/?from=unsafe#fragment"); err != nil {
		t.Fatal(err)
	}
	submitTestApplication(t, service)

	body := email.SentEmails[0].Body
	if !strings.Contains(body, "https://dock.example.com/base/api/applications/public/email-action?action=approve#token=") {
		t.Fatalf("email does not use configured public URL: %s", body)
	}
	if strings.Contains(body, "from=unsafe") || strings.Contains(body, "#fragment") {
		t.Fatal("public URL query or fragment leaked into action links")
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

func TestApplicationServiceListPublicServersIncludesContainerLoad(t *testing.T) {
	service, _, ssh := setupAppService(t)
	ssh.ExecuteCommandFn = func(_ string, _ int, _, _, _, command string) (string, error) {
		if strings.HasPrefix(command, "docker ps -a ") {
			return "api\tubuntu:22.04\tUp 2 hours\t0.0.0.0:20000->22/tcp\tabc123\nworker\tubuntu:22.04\tExited (0) 1 hour ago\t\tdef456", nil
		}
		return "", nil
	}

	servers, err := service.ListPublicServers()
	if err != nil {
		t.Fatal(err)
	}
	if len(servers) != 1 || !servers[0].LoadAvailable || servers[0].RunningContainers != 1 || servers[0].TotalContainers != 2 {
		t.Fatalf("unexpected public server load: %+v", servers)
	}
}

func TestApplicationServiceListPublicServersKeepsUnavailableLoad(t *testing.T) {
	service, _, ssh := setupAppService(t)
	ssh.ExecuteCommandFn = func(_ string, _ int, _, _, _, command string) (string, error) {
		if strings.HasPrefix(command, "docker ps -a ") {
			return "", errors.New("docker unavailable")
		}
		return "", nil
	}

	servers, err := service.ListPublicServers()
	if err != nil {
		t.Fatal(err)
	}
	if len(servers) != 1 || servers[0].LoadAvailable || servers[0].RunningContainers != 0 || servers[0].TotalContainers != 0 {
		t.Fatalf("unexpected unavailable public server load: %+v", servers)
	}
}

func TestApplicationServiceReject(t *testing.T) {
	service, email, _ := setupAppService(t)
	application := submitTestApplication(t, service)
	response, err := service.Reject(application.ID)
	if err != nil {
		t.Fatal(err)
	}
	if response.Status != "rejected" || email.AsyncCalls != 1 {
		t.Fatalf("unexpected rejection result: %+v, emails=%d", response, email.AsyncCalls)
	}
	if _, err := service.Reject(application.ID); !errors.Is(err, ErrApplicationNotPending) {
		t.Fatalf("expected ErrApplicationNotPending, got %v", err)
	}
}

func TestApplicationServiceIgnoreDoesNotSendEmail(t *testing.T) {
	service, email, _ := setupAppService(t)
	application := submitTestApplication(t, service)
	emailCallsBeforeIgnore := email.AsyncCalls

	response, err := service.Ignore(application.ID)
	if err != nil {
		t.Fatal(err)
	}
	if response.Status != "ignored" {
		t.Fatalf("unexpected ignore result: %+v", response)
	}
	if email.AsyncCalls != emailCallsBeforeIgnore {
		t.Fatalf("ignored application should not send email, calls changed from %d to %d", emailCallsBeforeIgnore, email.AsyncCalls)
	}
	if response.ConnectionInfo != nil {
		t.Fatal("ignored application should not include connection information")
	}
	if _, err := service.Ignore(application.ID); !errors.Is(err, ErrApplicationNotPending) {
		t.Fatalf("expected ErrApplicationNotPending, got %v", err)
	}
}

func TestApplicationServiceApprove(t *testing.T) {
	service, email, ssh := setupAppService(t)
	approvalTime := time.Date(2026, time.August, 28, 0, 1, 2, 0, time.FixedZone("CST", 8*60*60))
	service.now = func() time.Time { return approvalTime }
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
	if err := service.db.Model(&model.Application{}).Where("id = ?", application.ID).Update("applicant_name", "张三").Error; err != nil {
		t.Fatal(err)
	}

	response, err := service.Approve(application.ID)
	if err != nil {
		t.Fatal(err)
	}
	if response.Status != "approved" || email.AsyncCalls != 1 {
		t.Fatalf("unexpected approval result: %+v, emails=%d", response, email.AsyncCalls)
	}
	if response.ConnectionInfo == nil {
		t.Fatal("expected approval response to include connection information")
	}
	connection := response.ConnectionInfo
	if connection.Server != "10.0.0.1" || connection.User != "root" || connection.SSHPort == 0 || connection.Password == "" {
		t.Fatalf("unexpected connection information: %+v", connection)
	}
	if connection.ExtraPorts != "20001-20005" {
		t.Fatalf("unexpected extra ports: %q", connection.ExtraPorts)
	}
	expectedCommand := fmt.Sprintf("ssh -p %d root@10.0.0.1", connection.SSHPort)
	if connection.SSHCommand != expectedCommand {
		t.Fatalf("unexpected SSH command: %q", connection.SSHCommand)
	}
	if strings.ContainsAny(runCommand, "\r\n") {
		t.Fatalf("docker command should be one line: %q", runCommand)
	}
	containerName := "zhangsan-2026-08-28-00-01-02"
	for _, value := range []string{
		"--name " + containerName,
		"-v " + containerName + ":/data",
		"--gpus all",
		"--shm-size=8g",
		"nvidia/cuda:12.0",
	} {
		if !strings.Contains(runCommand, value) {
			t.Fatalf("docker command does not contain %q: %s", value, runCommand)
		}
	}
	if strings.Contains(runCommand, containerName+"-data") {
		t.Fatalf("volume should use the container name without a suffix: %s", runCommand)
	}
}

func TestApplicationServiceApproveContainerFailureLeavesApplicationRetryable(t *testing.T) {
	service, _, ssh := setupAppService(t)
	createShouldFail := true
	ssh.ExecuteCommandFn = func(_ string, _ int, _, _, _, command string) (string, error) {
		if strings.HasPrefix(command, "docker run ") && createShouldFail {
			return "", errors.New("docker daemon unavailable")
		}
		return "created", nil
	}
	application := submitTestApplication(t, service)
	_, err := service.Approve(application.ID)
	if !errors.Is(err, ErrContainerProvisioning) || !strings.Contains(err.Error(), "docker daemon unavailable") {
		t.Fatalf("unexpected error: %v", err)
	}
	var stored model.Application
	if err := service.db.First(&stored, application.ID).Error; err != nil {
		t.Fatal(err)
	}
	if stored.Status != "pending" {
		t.Fatalf("failed approval changed application status to %q", stored.Status)
	}

	createShouldFail = false
	response, err := service.Approve(application.ID)
	if err != nil {
		t.Fatalf("retry approval: %v", err)
	}
	if response.Status != "approved" {
		t.Fatalf("unexpected retry result: %+v", response)
	}
}

func TestApplicationServiceApproveRejectsEmptyImageAddress(t *testing.T) {
	service, _, _ := setupAppService(t)
	if err := service.db.Model(&model.Image{}).Where("id = ?", 1).Update("image_address", "").Error; err != nil {
		t.Fatal(err)
	}
	application := submitTestApplication(t, service)
	_, err := service.Approve(application.ID)
	if !errors.Is(err, ErrContainerProvisioning) || !strings.Contains(err.Error(), "empty image address") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApplicationServiceMissingImageDoesNotHidePendingApplications(t *testing.T) {
	service, _, _ := setupAppService(t)
	toReject := submitTestApplication(t, service)
	toIgnore := submitTestApplication(t, service)
	if err := service.db.Delete(&model.Image{}, toReject.ImageID).Error; err != nil {
		t.Fatal(err)
	}

	applications, err := service.List()
	if err != nil || len(applications) != 2 {
		t.Fatalf("expected orphaned applications to remain visible, got %d (%v)", len(applications), err)
	}
	if _, err := service.Approve(toReject.ID); !errors.Is(err, ErrContainerProvisioning) {
		t.Fatalf("expected missing image to be a provisioning error, got %v", err)
	}
	if response, err := service.Reject(toReject.ID); err != nil || response.Status != "rejected" {
		t.Fatalf("reject application with missing image: response=%+v err=%v", response, err)
	}
	if response, err := service.Ignore(toIgnore.ID); err != nil || response.Status != "ignored" {
		t.Fatalf("ignore application with missing image: response=%+v err=%v", response, err)
	}
}
