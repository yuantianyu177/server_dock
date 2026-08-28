package service

import (
	"strings"

	"serverdock/internal/dto"
)

func parseDockerImages(output string) []dto.RemoteImage {
	var images []dto.RemoteImage
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		parts := strings.SplitN(line, "\t", 5)
		if len(parts) != 5 {
			continue
		}
		images = append(images, dto.RemoteImage{
			Repository: parts[0], Tag: parts[1], ImageID: parts[2], Size: parts[3], Created: parts[4],
		})
	}
	return images
}

func parseDockerContainers(output string) []map[string]string {
	var containers []map[string]string
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		parts := strings.SplitN(line, "\t", 5)
		if len(parts) != 5 {
			continue
		}
		containers = append(containers, map[string]string{
			"name": parts[0], "image": parts[1], "status": parts[2], "ports": parts[3], "id": parts[4],
		})
	}
	return containers
}

func parseDockerVolumes(output string) []map[string]string {
	var volumes []map[string]string
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) != 3 {
			continue
		}
		volumes = append(volumes, map[string]string{
			"name": parts[0], "driver": parts[1], "mountpoint": parts[2],
		})
	}
	return volumes
}
