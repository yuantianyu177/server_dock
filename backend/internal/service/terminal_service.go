package service

import (
	"fmt"
	"io"
	"log/slog"
	"sync"

	"golang.org/x/crypto/ssh"
)

type TerminalSession struct {
	client  *ssh.Client
	session *ssh.Session
	stdin   io.WriteCloser
	stdout  io.Reader
	mu      sync.Mutex
}

type TerminalService struct {
	serverService *ServerService
}

func NewTerminalService(serverService *ServerService) *TerminalService {
	return &TerminalService{serverService: serverService}
}

// CreateSession creates an SSH terminal session to a server.
func (t *TerminalService) CreateSession(serverID uint, command string) (*TerminalSession, error) {
	server, credential, err := t.serverService.ResolveServer(serverID)
	if err != nil {
		return nil, err
	}
	client, err := dialSSH(server.Hostname, server.Port, server.User, server.AuthType, credential)
	if err != nil {
		return nil, fmt.Errorf("SSH connection failed: %w", err)
	}

	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	closeSession := func() {
		session.Close()
		client.Close()
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	if err := session.RequestPty("xterm-256color", 40, 120, modes); err != nil {
		closeSession()
		return nil, fmt.Errorf("failed to request PTY: %w", err)
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		closeSession()
		return nil, err
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		closeSession()
		return nil, err
	}

	var startErr error
	if command != "" {
		startErr = session.Start(command)
	} else {
		startErr = session.Shell()
	}
	if startErr != nil {
		closeSession()
		return nil, startErr
	}

	return &TerminalSession{
		client:  client,
		session: session,
		stdin:   stdin,
		stdout:  stdout,
	}, nil
}

func (ts *TerminalSession) Close() {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.stdin != nil {
		ts.stdin.Close()
	}
	if ts.session != nil {
		ts.session.Close()
	}
	if ts.client != nil {
		ts.client.Close()
	}
	slog.Debug("Terminal session closed")
}

func (ts *TerminalSession) Write(data []byte) (int, error) {
	return ts.stdin.Write(data)
}

func (ts *TerminalSession) Read(buf []byte) (int, error) {
	return ts.stdout.Read(buf)
}

// Resize sends a window resize request to the SSH session.
func (ts *TerminalSession) Resize(rows, cols int) error {
	return ts.session.WindowChange(rows, cols)
}
