package service

import (
	"fmt"
	"strings"

	"serverdock/internal/dto"
)

type RealDockerService struct {
	ssh SSHService
}

func NewRealDockerService(ssh SSHService) *RealDockerService {
	return &RealDockerService{ssh: ssh}
}

func (d *RealDockerService) ListImages(hostname string, port int, user, authType, credential string) ([]dto.RemoteImage, error) {
	output, err := d.ssh.ExecuteCommand(hostname, port, user, authType, credential,
		"docker images --format '{{.Repository}}\\t{{.Tag}}\\t{{.ID}}\\t{{.Size}}\\t{{.CreatedSince}}'")
	if err != nil {
		return nil, err
	}
	return ParseDockerImages(output), nil
}

func (d *RealDockerService) PullImage(hostname string, port int, user, authType, credential string, image string) (string, error) {
	return d.ssh.ExecuteCommand(hostname, port, user, authType, credential, fmt.Sprintf("docker pull '%s'", image))
}

func (d *RealDockerService) RemoveImage(hostname string, port int, user, authType, credential string, imageID string) error {
	_, err := d.ssh.ExecuteCommand(hostname, port, user, authType, credential, fmt.Sprintf("docker rmi '%s'", imageID))
	return err
}

func (d *RealDockerService) ListContainers(hostname string, port int, user, authType, credential string) ([]map[string]string, error) {
	output, err := d.ssh.ExecuteCommand(hostname, port, user, authType, credential,
		"docker ps -a --format '{{.Names}}\\t{{.Image}}\\t{{.Status}}\\t{{.Ports}}\\t{{.ID}}'")
	if err != nil {
		return nil, err
	}
	return ParseDockerContainers(output), nil
}

func (d *RealDockerService) ContainerAction(hostname string, port int, user, authType, credential string, name, action string) error {
	cmd := fmt.Sprintf("docker %s %s", action, name)
	_, err := d.ssh.ExecuteCommand(hostname, port, user, authType, credential, cmd)
	return err
}

func (d *RealDockerService) GetContainerLogs(hostname string, port int, user, authType, credential string, name string, tail int) (string, error) {
	cmd := fmt.Sprintf("docker logs --tail %d %s", tail, name)
	return d.ssh.ExecuteCommand(hostname, port, user, authType, credential, cmd)
}

func (d *RealDockerService) CreateContainer(hostname string, port int, user, authType, credential string, cmd string) (string, error) {
	return d.ssh.ExecuteCommand(hostname, port, user, authType, credential, cmd)
}

func (d *RealDockerService) ExecuteCommand(hostname string, port int, user, authType, credential string, command string) (string, error) {
	return d.ssh.ExecuteCommand(hostname, port, user, authType, credential, command)
}

func (d *RealDockerService) ListVolumes(hostname string, port int, user, authType, credential string) ([]map[string]string, error) {
	output, err := d.ssh.ExecuteCommand(hostname, port, user, authType, credential,
		"docker volume ls --format '{{.Name}}\\t{{.Driver}}\\t{{.Mountpoint}}'")
	if err != nil {
		return nil, err
	}
	return ParseDockerVolumes(output), nil
}

func (d *RealDockerService) CreateVolume(hostname string, port int, user, authType, credential string, name string) error {
	_, err := d.ssh.ExecuteCommand(hostname, port, user, authType, credential, fmt.Sprintf("docker volume create '%s'", name))
	return err
}

func (d *RealDockerService) RemoveVolume(hostname string, port int, user, authType, credential string, name string) error {
	_, err := d.ssh.ExecuteCommand(hostname, port, user, authType, credential, fmt.Sprintf("docker volume rm '%s'", name))
	return err
}

// ParseDockerImages parses `docker images` output.
func ParseDockerImages(output string) []dto.RemoteImage {
	var images []dto.RemoteImage
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 5)
		if len(parts) < 5 {
			continue
		}
		images = append(images, dto.RemoteImage{
			Repository: parts[0],
			Tag:        parts[1],
			ImageID:    parts[2],
			Size:       parts[3],
			Created:    parts[4],
		})
	}
	return images
}

// ParseDockerContainers parses `docker ps -a` output.
func ParseDockerContainers(output string) []map[string]string {
	var containers []map[string]string
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 5)
		if len(parts) < 5 {
			continue
		}
		containers = append(containers, map[string]string{
			"name":   parts[0],
			"image":  parts[1],
			"status": parts[2],
			"ports":  parts[3],
			"id":     parts[4],
		})
	}
	return containers
}

// ParseDockerVolumes parses `docker volume ls` output.
func ParseDockerVolumes(output string) []map[string]string {
	var volumes []map[string]string
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) < 3 {
			continue
		}
		volumes = append(volumes, map[string]string{
			"name":       parts[0],
			"driver":     parts[1],
			"mountpoint": parts[2],
		})
	}
	return volumes
}
