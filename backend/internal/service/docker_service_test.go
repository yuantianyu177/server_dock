package service

import (
	"testing"
)

func TestParseDockerImages(t *testing.T) {
	output := "ubuntu\t22.04\tabc123\t77.8MB\t2 weeks ago\nnginx\tlatest\tdef456\t142MB\t3 days ago"
	images := ParseDockerImages(output)

	if len(images) != 2 {
		t.Fatalf("Expected 2 images, got %d", len(images))
	}
	if images[0].Repository != "ubuntu" {
		t.Fatalf("Expected 'ubuntu', got %s", images[0].Repository)
	}
	if images[0].Tag != "22.04" {
		t.Fatalf("Expected '22.04', got %s", images[0].Tag)
	}
	if images[1].ImageID != "def456" {
		t.Fatalf("Expected 'def456', got %s", images[1].ImageID)
	}
}

func TestParseDockerImagesEmpty(t *testing.T) {
	images := ParseDockerImages("")
	if len(images) != 0 {
		t.Fatalf("Expected 0, got %d", len(images))
	}
}

func TestParseDockerContainers(t *testing.T) {
	output := "mycontainer\tubuntu:22.04\tUp 2 hours\t0.0.0.0:20000->22/tcp\tabc123"
	containers := ParseDockerContainers(output)

	if len(containers) != 1 {
		t.Fatalf("Expected 1, got %d", len(containers))
	}
	if containers[0]["name"] != "mycontainer" {
		t.Fatalf("Expected 'mycontainer', got %s", containers[0]["name"])
	}
}

func TestParseDockerVolumes(t *testing.T) {
	output := "myvolume\tlocal\t/var/lib/docker/volumes/myvolume/_data"
	volumes := ParseDockerVolumes(output)

	if len(volumes) != 1 {
		t.Fatalf("Expected 1, got %d", len(volumes))
	}
	if volumes[0]["name"] != "myvolume" {
		t.Fatalf("Expected 'myvolume', got %s", volumes[0]["name"])
	}
}
