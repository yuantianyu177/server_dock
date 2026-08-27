package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

var containerNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

type ContainerService struct{ serverService *ServerService }

func NewContainerService(serverService *ServerService) *ContainerService {
	return &ContainerService{serverService: serverService}
}

func ValidateContainerName(name string) bool { return containerNameRegex.MatchString(name) }

func parseUsedPorts(output string) map[int]bool {
	ports := make(map[int]bool)
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		if port, err := strconv.Atoi(fields[3][strings.LastIndex(fields[3], ":")+1:]); err == nil {
			ports[port] = true
		}
	}
	return ports
}

func parseDockerPorts(output string) map[int]bool {
	ports := make(map[int]bool)
	for _, part := range strings.FieldsFunc(output, func(r rune) bool { return r == ',' || r == '\n' }) {
		host, _, ok := strings.Cut(strings.TrimSpace(part), "->")
		if !ok {
			continue
		}
		if port, err := strconv.Atoi(host[strings.LastIndex(host, ":")+1:]); err == nil {
			ports[port] = true
		}
	}
	return ports
}

func allocatePorts(used map[int]bool, start, end, count int) ([]int, error) {
	ports := make([]int, 0, count)
	for port := start; port <= end && len(ports) < count; port++ {
		if !used[port] {
			ports = append(ports, port)
		}
	}
	if len(ports) != count {
		return nil, fmt.Errorf("not enough available ports: need %d, found %d", count, len(ports))
	}
	return ports, nil
}

func buildDockerRunCommand(name, image string, sshPort int, extraPorts []int, volumeName, mountPath, extraArgs string) string {
	parts := []string{"docker run -d", fmt.Sprintf("--name %s", name), fmt.Sprintf("-p %d:22", sshPort)}
	for _, port := range extraPorts {
		parts = append(parts, fmt.Sprintf("-p %d:%d", port, port))
	}
	if volumeName != "" && mountPath != "" {
		parts = append(parts, fmt.Sprintf("-v %s:%s", volumeName, mountPath))
	}
	if extraArgs = normalizeDockerExtraArgs(extraArgs); extraArgs != "" {
		parts = append(parts, extraArgs)
	}
	return strings.Join(append(parts, "--restart unless-stopped", image), " ")
}

func normalizeDockerExtraArgs(extraArgs string) string {
	return strings.Join(strings.Fields(extraArgs), " ")
}

func (s *ContainerService) CreateContainer(serverID uint, name, image, extraArgs string, portStart, portEnd, extraPortCount int, mountPath string) (map[string]interface{}, error) {
	if !ValidateContainerName(name) {
		return nil, errors.New("invalid container name: only [a-zA-Z0-9_-] allowed")
	}
	if image = strings.TrimSpace(image); image == "" {
		return nil, errors.New("image is required")
	}

	server, credential, err := s.serverService.ResolveServer(serverID)
	if err != nil {
		return nil, err
	}
	execute := func(command string) (string, error) {
		return s.serverService.runCommand(server.Hostname, server.Port, server.User, server.AuthType, credential, command)
	}

	var ssOutput, dockerOutput string
	var probes sync.WaitGroup
	probes.Add(2)
	go func() {
		defer probes.Done()
		ssOutput, _ = execute("ss -tlnp")
	}()
	go func() {
		defer probes.Done()
		dockerOutput, _ = execute("docker ps --format '{{.Ports}}'")
	}()
	probes.Wait()

	used := parseUsedPorts(ssOutput)
	for port := range parseDockerPorts(dockerOutput) {
		used[port] = true
	}
	ports, err := allocatePorts(used, portStart, portEnd, 1+extraPortCount)
	if err != nil {
		return nil, err
	}

	// Docker creates a same-named volume referenced by -v, avoiding another SSH round trip.
	volumeName := name
	password := rand.Text()
	command := buildDockerRunCommand(name, image, ports[0], ports[1:], volumeName, mountPath, extraArgs)
	output, err := execute(command)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}
	go execute(fmt.Sprintf("docker exec %s bash -c \"echo 'root:%s' | chpasswd\"", name, password))

	return map[string]interface{}{
		"name": name, "ssh_port": ports[0], "extra_ports": ports[1:],
		"volume": volumeName, "password": password, "output": output,
	}, nil
}

func (s *ContainerService) ListContainers(serverID uint) ([]map[string]string, error) {
	output, err := s.serverService.ExecuteCommand(serverID,
		"docker ps -a --format '{{.Names}}\\t{{.Image}}\\t{{.Status}}\\t{{.Ports}}\\t{{.ID}}'")
	if err != nil {
		return nil, err
	}
	return parseDockerContainers(output), nil
}

func (s *ContainerService) ContainerAction(serverID uint, name, action string) error {
	if !ValidateContainerName(name) {
		return errors.New("invalid container name")
	}
	if action == "delete" {
		action = "rm -f"
	}
	_, err := s.serverService.ExecuteCommand(serverID, fmt.Sprintf("docker %s %s", action, name))
	return err
}

func (s *ContainerService) GetContainerLogs(serverID uint, name string, tail int) (string, error) {
	if !ValidateContainerName(name) {
		return "", errors.New("invalid container name")
	}
	if tail <= 0 {
		tail = 100
	}
	return s.serverService.ExecuteCommand(serverID, fmt.Sprintf("docker logs --tail %d %s", tail, name))
}

func (s *ContainerService) ListVolumes(serverID uint) ([]map[string]string, error) {
	output, err := s.serverService.ExecuteCommand(serverID,
		"docker volume ls --format '{{.Name}}\\t{{.Driver}}\\t{{.Mountpoint}}'")
	if err != nil {
		return nil, err
	}
	return parseDockerVolumes(output), nil
}

func (s *ContainerService) CreateVolumeSingle(serverID uint, name string) error {
	_, err := s.serverService.ExecuteCommand(serverID, fmt.Sprintf("docker volume create '%s'", name))
	return err
}

func (s *ContainerService) RemoveVolume(serverID uint, name string) error {
	_, err := s.serverService.ExecuteCommand(serverID, fmt.Sprintf("docker volume rm '%s'", name))
	return err
}
