package service

import (
	"testing"
)

func TestValidateContainerName(t *testing.T) {
	tests := []struct {
		name  string
		valid bool
	}{
		{"my-container", true},
		{"my_container", true},
		{"MyContainer123", true},
		{"zhangsan-20260305143052", true},
		{"invalid name", false},
		{"bad;name", false},
		{"bad$name", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := ValidateContainerName(tt.name); got != tt.valid {
			t.Errorf("ValidateContainerName(%q) = %v, want %v", tt.name, got, tt.valid)
		}
	}
}

func TestValidateDockerCommand(t *testing.T) {
	tests := []struct {
		cmd   string
		valid bool
	}{
		{"docker ps", true},
		{"docker exec my-container ls", true},
		{"  docker images", true},
		{"rm -rf /", false},
		{"ls -la", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := ValidateDockerCommand(tt.cmd); got != tt.valid {
			t.Errorf("ValidateDockerCommand(%q) = %v, want %v", tt.cmd, got, tt.valid)
		}
	}
}

func TestParseUsedPorts(t *testing.T) {
	output := `State  Recv-Q Send-Q Local Address:Port  Peer Address:Port
LISTEN 0      128    0.0.0.0:22    0.0.0.0:*
LISTEN 0      128    0.0.0.0:80    0.0.0.0:*
LISTEN 0      128    [::]:443      [::]:*`

	ports := ParseUsedPorts(output)
	if !ports[22] {
		t.Error("Expected port 22 to be used")
	}
	if !ports[80] {
		t.Error("Expected port 80 to be used")
	}
	if !ports[443] {
		t.Error("Expected port 443 to be used")
	}
	if ports[8080] {
		t.Error("Port 8080 should not be used")
	}
}

func TestParseDockerPorts(t *testing.T) {
	output := `0.0.0.0:20000->22/tcp, 0.0.0.0:20001->20001/tcp
0.0.0.0:20010->22/tcp`

	ports := ParseDockerPorts(output)
	if !ports[20000] {
		t.Error("Expected port 20000")
	}
	if !ports[20001] {
		t.Error("Expected port 20001")
	}
	if !ports[20010] {
		t.Error("Expected port 20010")
	}
}

func TestAllocatePorts(t *testing.T) {
	used := map[int]bool{20000: true, 20001: true}

	ports, err := AllocatePorts(used, 20000, 20010, 3)
	if err != nil {
		t.Fatalf("AllocatePorts failed: %v", err)
	}
	if len(ports) != 3 {
		t.Fatalf("Expected 3 ports, got %d", len(ports))
	}
	// First available should be 20002
	if ports[0] != 20002 {
		t.Fatalf("Expected 20002, got %d", ports[0])
	}
}

func TestAllocatePortsNotEnough(t *testing.T) {
	used := map[int]bool{}
	for i := 20000; i <= 20005; i++ {
		used[i] = true
	}

	_, err := AllocatePorts(used, 20000, 20005, 1)
	if err == nil {
		t.Fatal("Expected error when no ports available")
	}
}

func TestBuildDockerRunCommand(t *testing.T) {
	cmd := BuildDockerRunCommand("mycontainer", "ubuntu:22.04", 20000, []int{20001, 20002}, "myvolume", "/data", "--gpus all")

	expected := []string{
		"docker run -d",
		"--name mycontainer",
		"-p 20000:22",
		"-p 20001:20001",
		"-p 20002:20002",
		"-v myvolume:/data",
		"--gpus all",
		"--restart unless-stopped",
		"ubuntu:22.04",
	}

	for _, part := range expected {
		if !contains(cmd, part) {
			t.Errorf("Command missing %q: %s", part, cmd)
		}
	}
}

func TestBuildDockerRunCommand_NormalizesMultilineExtraArgs(t *testing.T) {
	cmd := BuildDockerRunCommand(
		"mycontainer",
		"ubuntu:22.04",
		20000,
		nil,
		"myvolume",
		"/data",
		"\n  --gpus all\r\n--shm-size=8g  \n\n",
	)

	if contains(cmd, "\n") || contains(cmd, "\r") {
		t.Fatalf("expected normalized command without newlines, got %q", cmd)
	}
	for _, part := range []string{"--gpus all", "--shm-size=8g", "ubuntu:22.04"} {
		if !contains(cmd, part) {
			t.Errorf("Command missing %q: %s", part, cmd)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsStr(s, substr)
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
