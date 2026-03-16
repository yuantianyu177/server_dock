package service

import (
	"fmt"
	"time"

	"golang.org/x/crypto/ssh"
)

type RealSSHService struct{}

func NewRealSSHService() *RealSSHService {
	return &RealSSHService{}
}

func (s *RealSSHService) TestConnection(hostname string, port int, user, authType, credential string) error {
	client, err := s.connect(hostname, port, user, authType, credential)
	if err != nil {
		return err
	}
	client.Close()
	return nil
}

func (s *RealSSHService) ExecuteCommand(hostname string, port int, user, authType, credential string, command string) (string, error) {
	client, err := s.connect(hostname, port, user, authType, credential)
	if err != nil {
		return "", err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		return string(output), fmt.Errorf("command failed: %w, output: %s", err, string(output))
	}
	return string(output), nil
}

func (s *RealSSHService) connect(hostname string, port int, user, authType, credential string) (*ssh.Client, error) {
	var authMethods []ssh.AuthMethod

	switch authType {
	case "password":
		authMethods = []ssh.AuthMethod{ssh.Password(credential)}
	case "key":
		signer, err := ssh.ParsePrivateKey([]byte(credential))
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		authMethods = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	default:
		return nil, fmt.Errorf("unsupported auth type: %s", authType)
	}

	config := &ssh.ClientConfig{
		User:            user,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", hostname, port)
	return ssh.Dial("tcp", addr, config)
}
