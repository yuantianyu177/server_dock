package service

import "serverdock/internal/dto"

// DockerService defines the interface for Docker operations via SSH.
type DockerService interface {
	ListImages(hostname string, port int, user, authType, credential string) ([]dto.RemoteImage, error)
	PullImage(hostname string, port int, user, authType, credential string, image string) (string, error)
	RemoveImage(hostname string, port int, user, authType, credential string, imageID string) error
	ListContainers(hostname string, port int, user, authType, credential string) ([]map[string]string, error)
	ContainerAction(hostname string, port int, user, authType, credential string, name, action string) error
	GetContainerLogs(hostname string, port int, user, authType, credential string, name string, tail int) (string, error)
	CreateContainer(hostname string, port int, user, authType, credential string, cmd string) (string, error)
	ExecuteCommand(hostname string, port int, user, authType, credential string, command string) (string, error)
	ListVolumes(hostname string, port int, user, authType, credential string) ([]map[string]string, error)
	CreateVolume(hostname string, port int, user, authType, credential string, name string) error
	RemoveVolume(hostname string, port int, user, authType, credential string, name string) error
}
