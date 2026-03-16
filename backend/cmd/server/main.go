package main

import (
	"log/slog"
	"os"

	"serverdock/internal/config"
	"serverdock/internal/model"
	"serverdock/internal/repository"
	"serverdock/internal/router"
	"serverdock/internal/service"
)

func main() {
	cfg := config.Load()

	// Setup logger
	logLevel := slog.LevelInfo
	if cfg.Debug {
		logLevel = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})))

	// Initialize database
	db, err := model.InitDB(cfg.DatabaseURL)
	if err != nil {
		slog.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}

	// Initialize repositories
	adminRepo := repository.NewAdminRepo(db)
	serverRepo := repository.NewServerRepo(db)
	imageRepo := repository.NewImageRepo(db)
	configRepo := repository.NewConfigRepo(db)
	appRepo := repository.NewApplicationRepo(db)

	// Initialize services
	sshService := service.NewRealSSHService()
	dockerService := service.NewRealDockerService(sshService)

	authService := service.NewAuthService(adminRepo, cfg.SecretKey)
	serverService := service.NewServerService(serverRepo, sshService, cfg.SSHCredentialKey)
	imageService := service.NewImageService(imageRepo, serverService, dockerService)
	containerService := service.NewContainerService(serverService, dockerService, sshService)
	configService := service.NewConfigService(configRepo)
	emailService := service.NewSMTPEmailService(configService)
	appService := service.NewApplicationService(appRepo, imageRepo, serverService, containerService, configService, emailService)
	terminalService := service.NewTerminalService(serverService)

	// Ensure default configs
	if err := configService.EnsureDefaults(); err != nil {
		slog.Error("Failed to initialize default configs", "error", err)
		os.Exit(1)
	}

	// Ensure default admin exists
	if err := authService.EnsureDefaultAdmin(cfg.DefaultAdminUsername, cfg.DefaultAdminPassword); err != nil {
		slog.Error("Failed to create default admin", "error", err)
		os.Exit(1)
	}

	// Setup router
	r := router.Setup(cfg.Debug, &router.Services{
		Auth:      authService,
		Server:    serverService,
		Image:     imageService,
		Container: containerService,
		Config:    configService,
		Email:       emailService,
		Application: appService,
		Terminal:    terminalService,
	})

	slog.Info("Server starting", "port", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		slog.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}
