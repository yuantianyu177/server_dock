package service

import (
	"fmt"
	"time"

	"golang.org/x/crypto/ssh"
)

type SSHTestFunc func(hostname string, port int, user, authType, credential string) error
type SSHRunFunc func(hostname string, port int, user, authType, credential, command string) (string, error)

func TestSSHConnection(hostname string, port int, user, authType, credential string) error {
	client, err := dialSSH(hostname, port, user, authType, credential)
	if err == nil {
		client.Close()
	}
	return err
}

func ExecuteSSHCommand(hostname string, port int, user, authType, credential, command string) (string, error) {
	client, err := dialSSH(hostname, port, user, authType, credential)
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
		return string(output), fmt.Errorf("command failed: %w, output: %s", err, output)
	}
	return string(output), nil
}

func dialSSH(hostname string, port int, user, authType, credential string) (*ssh.Client, error) {
	var auth ssh.AuthMethod
	switch authType {
	case "password":
		auth = ssh.Password(credential)
	case "key":
		signer, err := ssh.ParsePrivateKey([]byte(credential))
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		auth = ssh.PublicKeys(signer)
	default:
		return nil, fmt.Errorf("unsupported auth type: %s", authType)
	}

	return ssh.Dial("tcp", fmt.Sprintf("%s:%d", hostname, port), &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{auth},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	})
}
