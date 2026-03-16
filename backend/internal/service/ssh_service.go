package service

// SSHService defines the interface for SSH operations.
type SSHService interface {
	// TestConnection tests if an SSH connection can be established.
	TestConnection(hostname string, port int, user, authType, credential string) error
	// ExecuteCommand runs a command on a remote server and returns stdout.
	ExecuteCommand(hostname string, port int, user, authType, credential string, command string) (string, error)
}
