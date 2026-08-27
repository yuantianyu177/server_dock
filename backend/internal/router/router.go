package router

import (
	"serverdock/internal/handler"
	"serverdock/internal/service"

	"github.com/gin-gonic/gin"
)

type Services struct {
	Auth        *service.AuthService
	Server      *service.ServerService
	Image       *service.ImageService
	Container   *service.ContainerService
	Config      *service.ConfigService
	Application *service.ApplicationService
	Terminal    *service.TerminalService
	SendEmail   func(string, string, string) error
}

func Setup(debug bool, services *Services) *gin.Engine {
	if !debug {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()

	auth := handler.NewAuthHandler(services.Auth)
	servers := handler.NewServerHandler(services.Server)
	images := handler.NewImageHandler(services.Image)
	containers := handler.NewContainerHandler(services.Container)
	volumes := handler.NewVolumeHandler(services.Container)
	config := handler.NewConfigHandler(services.Config, services.SendEmail)
	applications := handler.NewApplicationHandler(services.Application)
	terminal := handler.NewTerminalHandler(services.Terminal, services.Auth)

	api := router.Group("/api")
	api.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	api.POST("/auth/login", auth.Login)

	public := api.Group("/applications/public")
	public.GET("/servers", applications.PublicListServers)
	public.GET("/server/:id/images", applications.PublicListImages)
	public.POST("/apply", applications.PublicSubmit)

	authed := api.Group("")
	authed.Use(handler.AuthMiddleware(services.Auth))
	authed.GET("/auth/me", auth.Me)
	authed.POST("/auth/change-password", auth.ChangePassword)

	authed.GET("/servers", servers.List)
	authed.POST("/servers", servers.Create)
	authed.GET("/servers/:id", servers.Get)
	authed.PUT("/servers/:id", servers.Update)
	authed.DELETE("/servers/:id", servers.Delete)
	authed.POST("/servers/:id/test", servers.TestConnection)
	authed.POST("/servers/test-direct", servers.TestConnectionDirect)

	authed.GET("/servers/:id/images", images.ListRemote)
	authed.POST("/servers/:id/images/pull", images.PullRemote)
	authed.DELETE("/servers/:id/images/:image_id", images.RemoveRemote)
	authed.GET("/images", images.List)
	authed.POST("/images", images.Create)
	authed.DELETE("/images/:id", images.Delete)

	authed.GET("/servers/:id/containers", containers.List)
	authed.POST("/servers/:id/containers", containers.Create)
	authed.POST("/servers/:id/containers/:name/action", containers.Action)
	authed.GET("/servers/:id/containers/:name/logs", containers.Logs)
	authed.GET("/servers/:id/volumes", volumes.List)
	authed.POST("/servers/:id/volumes", volumes.Create)
	authed.DELETE("/servers/:id/volumes/:name", volumes.Delete)

	authed.GET("/config", config.GetAll)
	authed.PUT("/config/:key", config.Update)
	authed.POST("/config/test-email", config.TestEmail)
	authed.GET("/applications", applications.List)
	authed.POST("/applications/:id/action", applications.Action)

	api.GET("/terminal/ws/:server_id", terminal.ServerTerminal)
	api.GET("/terminal/container/ws/:server_id/:container_name", terminal.ContainerTerminal)
	return router
}
