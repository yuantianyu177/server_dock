package router

import (
	"serverdock/internal/handler"
	"serverdock/internal/pkg"
	"serverdock/internal/service"

	"github.com/gin-gonic/gin"
)

type Services struct {
	Auth      *service.AuthService
	Server    *service.ServerService
	Image     *service.ImageService
	Container *service.ContainerService
	Config      *service.ConfigService
	Email       service.EmailService
	Application *service.ApplicationService
	Terminal    *service.TerminalService
}

// Setup creates and configures the Gin router.
func Setup(debug bool, svc *Services) *gin.Engine {
	if !debug {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	authHandler := handler.NewAuthHandler(svc.Auth)
	serverHandler := handler.NewServerHandler(svc.Server)
	imageHandler := handler.NewImageHandler(svc.Image)
	containerHandler := handler.NewContainerHandler(svc.Container)
	volumeHandler := handler.NewVolumeHandler(svc.Container)
	configHandler := handler.NewConfigHandler(svc.Config, svc.Email)
	appHandler := handler.NewApplicationHandler(svc.Application)
	terminalHandler := handler.NewTerminalHandler(svc.Terminal, svc.Auth)

	api := r.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, pkg.SuccessResponse(gin.H{"status": "ok"}))
		})

		// Auth
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			protected := auth.Group("")
			protected.Use(handler.AuthMiddleware(svc.Auth))
			{
				protected.GET("/me", authHandler.Me)
				protected.POST("/change-password", authHandler.ChangePassword)
			}
		}

		// Public application endpoints
		appPublic := api.Group("/applications/public")
		{
			appPublic.GET("/servers", appHandler.PublicListServers)
			appPublic.GET("/server/:id/images", appHandler.PublicListImages)
			appPublic.POST("/apply", appHandler.PublicSubmit)
		}

		// Protected routes
		authed := api.Group("")
		authed.Use(handler.AuthMiddleware(svc.Auth))
		{
			// Servers
			servers := authed.Group("/servers")
			{
				servers.GET("", serverHandler.List)
				servers.POST("", serverHandler.Create)
				servers.GET("/:id", serverHandler.Get)
				servers.PUT("/:id", serverHandler.Update)
				servers.DELETE("/:id", serverHandler.Delete)
				servers.POST("/:id/test", serverHandler.TestConnection)
				servers.POST("/test-direct", serverHandler.TestConnectionDirect)

				// Remote images on server
				servers.GET("/:id/images", imageHandler.ListRemote)
				servers.POST("/:id/images/pull", imageHandler.PullRemote)
				servers.DELETE("/:id/images/:image_id", imageHandler.RemoveRemote)

				// Containers on server
				servers.GET("/:id/containers", containerHandler.List)
				servers.POST("/:id/containers", containerHandler.Create)
				servers.POST("/:id/containers/:name/action", containerHandler.Action)
				servers.GET("/:id/containers/:name/logs", containerHandler.Logs)
				servers.POST("/:id/exec", containerHandler.Exec)

				// Volumes on server
				servers.GET("/:id/volumes", volumeHandler.List)
				servers.POST("/:id/volumes", volumeHandler.Create)
				servers.DELETE("/:id/volumes/:name", volumeHandler.Delete)
			}

			// Images (DB records)
			images := authed.Group("/images")
			{
				images.GET("", imageHandler.List)
				images.POST("", imageHandler.Create)
				images.GET("/:id", imageHandler.Get)
				images.PUT("/:id", imageHandler.Update)
				images.DELETE("/:id", imageHandler.Delete)
			}

			// Config
			cfgGroup := authed.Group("/config")
			{
				cfgGroup.GET("", configHandler.List)
				cfgGroup.GET("/all", configHandler.GetAll)
				cfgGroup.PUT("/:key", configHandler.Update)
				cfgGroup.POST("/test-email", configHandler.TestEmail)
			}

			// Applications (admin)
			apps := authed.Group("/applications")
			{
				apps.GET("", appHandler.List)
				apps.GET("/:id", appHandler.Get)
				apps.POST("/:id/action", appHandler.Action)
			}
		}

		// Terminal (WebSocket, auth via query param)
		terminal := api.Group("/terminal")
		{
			terminal.GET("/ws/:server_id", terminalHandler.ServerTerminal)
			terminal.GET("/container/ws/:server_id/:container_name", terminalHandler.ContainerTerminal)
		}
	}

	return r
}
