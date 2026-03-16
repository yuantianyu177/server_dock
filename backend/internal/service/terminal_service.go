package service

import (
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"

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
	sessions      map[string]*TerminalSession
	mu            sync.RWMutex
}

func NewTerminalService(serverService *ServerService) *TerminalService {
	return &TerminalService{
		serverService: serverService,
		sessions:      make(map[string]*TerminalSession),
	}
}

// CreateSession creates an SSH terminal session to a server.
func (t *TerminalService) CreateSession(sessionID string, serverID uint, command string) (*TerminalSession, error) {
	server, err := t.serverService.GetRawByID(serverID)
	if err != nil {
		return nil, fmt.Errorf("server not found")
	}

	cred, err := t.serverService.DecryptCredential(server)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt credential")
	}

	var authMethods []ssh.AuthMethod
	switch server.AuthType {
	case "password":
		authMethods = []ssh.AuthMethod{ssh.Password(cred)}
	case "key":
		signer, err := ssh.ParsePrivateKey([]byte(cred))
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		authMethods = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	default:
		return nil, fmt.Errorf("unsupported auth type: %s", server.AuthType)
	}

	config := &ssh.ClientConfig{
		User:            server.User,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", server.Hostname, server.Port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("SSH connection failed: %w", err)
	}

	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Request PTY
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	if err := session.RequestPty("xterm-256color", 40, 120, modes); err != nil {
		session.Close()
		client.Close()
		return nil, fmt.Errorf("failed to request PTY: %w", err)
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		session.Close()
		client.Close()
		return nil, err
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		session.Close()
		client.Close()
		return nil, err
	}

	// Start shell or command
	if command != "" {
		if err := session.Start(command); err != nil {
			session.Close()
			client.Close()
			return nil, err
		}
	} else {
		if err := session.Shell(); err != nil {
			session.Close()
			client.Close()
			return nil, err
		}
	}

	ts := &TerminalSession{
		client:  client,
		session: session,
		stdin:   stdin,
		stdout:  stdout,
	}

	t.mu.Lock()
	t.sessions[sessionID] = ts
	t.mu.Unlock()

	return ts, nil
}

// CloseSession closes and cleans up a terminal session.
func (t *TerminalService) CloseSession(sessionID string) {
	t.mu.Lock()
	ts, ok := t.sessions[sessionID]
	if ok {
		delete(t.sessions, sessionID)
	}
	t.mu.Unlock()

	if ok && ts != nil {
		ts.Close()
	}
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
