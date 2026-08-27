package main

import (
	"log/slog"
	"os"

	"serverdock/internal/config"
	"serverdock/internal/model"
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

	authService := service.NewAuthService(db, cfg.SecretKey)
	serverService := service.NewServerService(db, service.TestSSHConnection, service.ExecuteSSHCommand, cfg.SSHCredentialKey)
	imageService := service.NewImageService(db, serverService)
	containerService := service.NewContainerService(serverService)
	configService := service.NewConfigService(db)
	emailService := service.NewSMTPEmailService(configService)
	appService := service.NewApplicationService(
		db,
		serverService,
		containerService,
		configService,
		cfg.SecretKey,
		cfg.PublicURL,
		emailService.SendAsync,
	)
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
		Auth:        authService,
		Server:      serverService,
		Image:       imageService,
		Container:   containerService,
		Config:      configService,
		Application: appService,
		Terminal:    terminalService,
		SendEmail:   emailService.Send,
	})

	slog.Info("Server starting", "port", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		slog.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}
