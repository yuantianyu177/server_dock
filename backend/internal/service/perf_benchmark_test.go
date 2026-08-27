package service

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"serverdock/internal/dto"
	"serverdock/internal/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func BenchmarkCreateContainerSSHLatency(b *testing.B) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		b.Fatal(err)
	}
	if err := db.AutoMigrate(&model.Server{}); err != nil {
		b.Fatal(err)
	}

	run := func(_ string, _ int, _, _, _, command string) (string, error) {
		if !strings.HasPrefix(command, "docker exec ") {
			time.Sleep(20 * time.Millisecond)
		}
		return "", nil
	}
	servers := NewServerService(db, nil, run, testEncryptKey)
	server, err := servers.Create(&dto.CreateServerRequest{
		Host: "Benchmark", Hostname: "127.0.0.1", User: "root", AuthType: "password", Credential: "secret",
	})
	if err != nil {
		b.Fatal(err)
	}
	containers := NewContainerService(servers)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := containers.CreateContainer(server.ID, fmt.Sprintf("bench-%d", i), "ubuntu:latest", "", 20000, 30000, 5, "/data"); err != nil {
			b.Fatal(err)
		}
	}
}
