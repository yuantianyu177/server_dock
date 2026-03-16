package service

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"serverdock/internal/pkg"
)

var containerNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

type ContainerService struct {
	serverService *ServerService
	dockerService DockerService
	sshService    SSHService
}

func NewContainerService(serverService *ServerService, dockerService DockerService, sshService SSHService) *ContainerService {
	return &ContainerService{serverService: serverService, dockerService: dockerService, sshService: sshService}
}

// ValidateContainerName checks if the name matches [a-zA-Z0-9_-].
func ValidateContainerName(name string) bool {
	return containerNameRegex.MatchString(name)
}

// ValidateDockerCommand ensures the command starts with "docker".
func ValidateDockerCommand(cmd string) bool {
	return strings.HasPrefix(strings.TrimSpace(cmd), "docker")
}

// ParseUsedPorts parses `ss -tlnp` output to extract listening ports.
func ParseUsedPorts(ssOutput string) map[int]bool {
	ports := make(map[int]bool)
	for _, line := range strings.Split(ssOutput, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "State") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		addr := fields[3]
		idx := strings.LastIndex(addr, ":")
		if idx < 0 {
			continue
		}
		portStr := addr[idx+1:]
		port, err := strconv.Atoi(portStr)
		if err == nil {
			ports[port] = true
		}
	}
	return ports
}

// ParseDockerPorts parses `docker ps` port mappings to extract host ports.
func ParseDockerPorts(dockerPsOutput string) map[int]bool {
	ports := make(map[int]bool)
	for _, line := range strings.Split(dockerPsOutput, "\n") {
		// Look for patterns like 0.0.0.0:20000->22/tcp
		for _, part := range strings.Split(line, ",") {
			part = strings.TrimSpace(part)
			if idx := strings.Index(part, "->"); idx > 0 {
				hostPart := part[:idx]
				colonIdx := strings.LastIndex(hostPart, ":")
				if colonIdx >= 0 {
					portStr := hostPart[colonIdx+1:]
					port, err := strconv.Atoi(portStr)
					if err == nil {
						ports[port] = true
					}
				}
			}
		}
	}
	return ports
}

// AllocatePorts finds N available ports in the given range.
func AllocatePorts(usedPorts map[int]bool, start, end, count int) ([]int, error) {
	var allocated []int
	for port := start; port <= end && len(allocated) < count; port++ {
		if !usedPorts[port] {
			allocated = append(allocated, port)
		}
	}
	if len(allocated) < count {
		return nil, fmt.Errorf("not enough available ports: need %d, found %d", count, len(allocated))
	}
	return allocated, nil
}

// BuildDockerRunCommand builds a `docker run` command for container creation.
func BuildDockerRunCommand(name, image string, sshPort int, extraPorts []int, volumeName, mountPath, extraArgs string) string {
	password := pkg.GenerateRandomPassword(16)
	cmd, _ := BuildDockerRunCommandWithPassword(name, image, sshPort, extraPorts, volumeName, mountPath, extraArgs, password)
	return cmd
}

func BuildDockerRunCommandWithPassword(name, image string, sshPort int, extraPorts []int, volumeName, mountPath, extraArgs, password string) (string, string) {
	parts := []string{"docker run -d", fmt.Sprintf("--name %s", name)}

	// SSH port mapping
	parts = append(parts, fmt.Sprintf("-p %d:22", sshPort))

	// Extra port mappings (same-number)
	for _, p := range extraPorts {
		parts = append(parts, fmt.Sprintf("-p %d:%d", p, p))
	}

	// Volume
	if volumeName != "" && mountPath != "" {
		parts = append(parts, fmt.Sprintf("-v %s:%s", volumeName, mountPath))
	}

	// Extra args
	if normalizedExtraArgs := NormalizeDockerExtraArgs(extraArgs); normalizedExtraArgs != "" {
		parts = append(parts, normalizedExtraArgs)
	}

	parts = append(parts, "--restart unless-stopped", image)

	return strings.Join(parts, " "), password
}

// NormalizeDockerExtraArgs collapses multiline config input into a single shell command fragment.
func NormalizeDockerExtraArgs(extraArgs string) string {
	if extraArgs == "" {
		return ""
	}

	lines := strings.Split(strings.ReplaceAll(extraArgs, "\r\n", "\n"), "\n")
	parts := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts = append(parts, line)
	}

	return strings.Join(parts, " ")
}


// CreateContainer creates a container on a server with port allocation.
func (s *ContainerService) CreateContainer(serverID uint, name, image, extraArgs string, portRangeStart, portRangeEnd, extraPortCount int, volumeMountPath string) (map[string]interface{}, error) {
	if !ValidateContainerName(name) {
		return nil, errors.New("invalid container name: only [a-zA-Z0-9_-] allowed")
	}
	image = strings.TrimSpace(image)
	if image == "" {
		return nil, errors.New("image is required")
	}

	server, cred, err := s.serverService.ResolveServer(serverID)
	if err != nil {
		return nil, err
	}

	// Get used ports (run ss and docker ps in parallel)
	type cmdResult struct {
		output string
		err    error
	}
	ssCh := make(chan cmdResult, 1)
	dpCh := make(chan cmdResult, 1)
	go func() {
		out, err := s.sshService.ExecuteCommand(server.Hostname, server.Port, server.User, server.AuthType, cred, "ss -tlnp")
		ssCh <- cmdResult{out, err}
	}()
	go func() {
		out, err := s.sshService.ExecuteCommand(server.Hostname, server.Port, server.User, server.AuthType, cred,
			"docker ps --format '{{.Ports}}'")
		dpCh <- cmdResult{out, err}
	}()
	ssResult := <-ssCh
	dpResult := <-dpCh

	usedPorts := ParseUsedPorts(ssResult.output)
	dockerPorts := ParseDockerPorts(dpResult.output)
	for p := range dockerPorts {
		usedPorts[p] = true
	}

	// Allocate ports: 1 SSH + N extra
	totalPorts := 1 + extraPortCount
	ports, err := AllocatePorts(usedPorts, portRangeStart, portRangeEnd, totalPorts)
	if err != nil {
		return nil, err
	}

	sshPort := ports[0]
	extraPorts := ports[1:]

	// Create volume
	volumeName := name + "-data"
	if err := s.dockerService.CreateVolume(server.Hostname, server.Port, server.User, server.AuthType, cred, volumeName); err != nil {
		return nil, fmt.Errorf("failed to create volume: %w", err)
	}

	// Build and run docker command
	password := pkg.GenerateRandomPassword(16)
	cmd, password := BuildDockerRunCommandWithPassword(name, image, sshPort, extraPorts, volumeName, volumeMountPath, extraArgs, password)
	output, err := s.dockerService.CreateContainer(server.Hostname, server.Port, server.User, server.AuthType, cred, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	// Set root password asynchronously after container starts
	go func() {
		chpasswdCmd := fmt.Sprintf("docker exec %s bash -c \"echo 'root:%s' | chpasswd\"", name, password)
		s.sshService.ExecuteCommand(server.Hostname, server.Port, server.User, server.AuthType, cred, chpasswdCmd)
	}()

	result := map[string]interface{}{
		"name":        name,
		"ssh_port":    sshPort,
		"extra_ports": extraPorts,
		"volume":      volumeName,
		"password":    password,
		"output":      output,
	}

	return result, nil
}

// ListContainers lists containers on a server.
func (s *ContainerService) ListContainers(serverID uint) ([]map[string]string, error) {
	server, cred, err := s.serverService.ResolveServer(serverID)
	if err != nil {
		return nil, err
	}

	return s.dockerService.ListContainers(server.Hostname, server.Port, server.User, server.AuthType, cred)
}

// ContainerAction performs start/stop/restart/delete on a container.
func (s *ContainerService) ContainerAction(serverID uint, name, action string) error {
	if !ValidateContainerName(name) {
		return errors.New("invalid container name")
	}

	server, cred, err := s.serverService.ResolveServer(serverID)
	if err != nil {
		return err
	}

	dockerAction := action
	if action == "delete" {
		// Force remove (stops and removes in one step)
		dockerAction = "rm -f"
	}

	return s.dockerService.ContainerAction(server.Hostname, server.Port, server.User, server.AuthType, cred, name, dockerAction)
}

// GetContainerLogs gets logs from a container.
func (s *ContainerService) GetContainerLogs(serverID uint, name string, tail int) (string, error) {
	if !ValidateContainerName(name) {
		return "", errors.New("invalid container name")
	}
	if tail <= 0 {
		tail = 100
	}

	server, cred, err := s.serverService.ResolveServer(serverID)
	if err != nil {
		return "", err
	}

	return s.dockerService.GetContainerLogs(server.Hostname, server.Port, server.User, server.AuthType, cred, name, tail)
}

// ExecCommand executes a docker command on a server.
func (s *ContainerService) ExecCommand(serverID uint, command string) (string, error) {
	if !ValidateDockerCommand(command) {
		return "", errors.New("only docker commands are allowed")
	}

	server, cred, err := s.serverService.ResolveServer(serverID)
	if err != nil {
		return "", err
	}

	return s.dockerService.ExecuteCommand(server.Hostname, server.Port, server.User, server.AuthType, cred, command)
}

// ListVolumes lists volumes on a server.
func (s *ContainerService) ListVolumes(serverID uint) ([]map[string]string, error) {
	server, cred, err := s.serverService.ResolveServer(serverID)
	if err != nil {
		return nil, err
	}

	return s.dockerService.ListVolumes(server.Hostname, server.Port, server.User, server.AuthType, cred)
}

// CreateVolumeSingle creates a volume on a server.
func (s *ContainerService) CreateVolumeSingle(serverID uint, name string) error {
	server, cred, err := s.serverService.ResolveServer(serverID)
	if err != nil {
		return err
	}

	return s.dockerService.CreateVolume(server.Hostname, server.Port, server.User, server.AuthType, cred, name)
}

// RemoveVolume removes a volume on a server.
func (s *ContainerService) RemoveVolume(serverID uint, name string) error {
	server, cred, err := s.serverService.ResolveServer(serverID)
	if err != nil {
		return err
	}

	return s.dockerService.RemoveVolume(server.Hostname, server.Port, server.User, server.AuthType, cred, name)
}
