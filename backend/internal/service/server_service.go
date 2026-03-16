package service

import (
	"errors"

	"serverdock/internal/dto"
	"serverdock/internal/model"
	"serverdock/internal/pkg"
	"serverdock/internal/repository"
)

type ServerService struct {
	serverRepo *repository.ServerRepo
	sshService SSHService
	encryptKey string
}

func NewServerService(serverRepo *repository.ServerRepo, sshService SSHService, encryptKey string) *ServerService {
	return &ServerService{serverRepo: serverRepo, sshService: sshService, encryptKey: encryptKey}
}

func (s *ServerService) Create(req *dto.CreateServerRequest) (*dto.ServerResponse, error) {
	port := req.Port
	if port == 0 {
		port = 22
	}

	encrypted, err := pkg.Encrypt(req.Credential, s.encryptKey)
	if err != nil {
		return nil, errors.New("failed to encrypt credential")
	}

	server := &model.Server{
		Host:        req.Host,
		Hostname:    req.Hostname,
		Port:        port,
		User:        req.User,
		AuthType:    req.AuthType,
		Credential:  encrypted,
		Description: req.Description,
	}

	if err := s.serverRepo.Create(server); err != nil {
		return nil, err
	}
	return s.toResponse(server), nil
}

func (s *ServerService) GetByID(id uint) (*dto.ServerResponse, error) {
	server, err := s.serverRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("server not found")
	}
	return s.toResponse(server), nil
}

func (s *ServerService) GetRawByID(id uint) (*model.Server, error) {
	return s.serverRepo.FindByID(id)
}

func (s *ServerService) List() ([]dto.ServerResponse, error) {
	servers, err := s.serverRepo.List()
	if err != nil {
		return nil, err
	}

	var responses []dto.ServerResponse
	for _, srv := range servers {
		responses = append(responses, *s.toResponse(&srv))
	}
	return responses, nil
}

func (s *ServerService) Update(id uint, req *dto.UpdateServerRequest) (*dto.ServerResponse, error) {
	server, err := s.serverRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("server not found")
	}

	if req.Host != "" {
		server.Host = req.Host
	}
	if req.Hostname != "" {
		server.Hostname = req.Hostname
	}
	if req.Port != 0 {
		server.Port = req.Port
	}
	if req.User != "" {
		server.User = req.User
	}
	if req.AuthType != "" {
		server.AuthType = req.AuthType
	}
	if req.Credential != "" {
		encrypted, err := pkg.Encrypt(req.Credential, s.encryptKey)
		if err != nil {
			return nil, errors.New("failed to encrypt credential")
		}
		server.Credential = encrypted
	}
	if req.Description != "" {
		server.Description = req.Description
	}

	if err := s.serverRepo.Update(server); err != nil {
		return nil, err
	}
	return s.toResponse(server), nil
}

func (s *ServerService) Delete(id uint) error {
	_, err := s.serverRepo.FindByID(id)
	if err != nil {
		return errors.New("server not found")
	}
	return s.serverRepo.Delete(id)
}

func (s *ServerService) TestConnectionDirect(hostname string, port int, user, authType, credential string) error {
	if port <= 0 {
		port = 22
	}
	return s.sshService.TestConnection(hostname, port, user, authType, credential)
}

func (s *ServerService) TestConnection(id uint) error {
	server, err := s.serverRepo.FindByID(id)
	if err != nil {
		return errors.New("server not found")
	}

	credential, err := pkg.Decrypt(server.Credential, s.encryptKey)
	if err != nil {
		return errors.New("failed to decrypt credential")
	}

	return s.sshService.TestConnection(server.Hostname, server.Port, server.User, server.AuthType, credential)
}

// DecryptCredential decrypts and returns the server's credential.
func (s *ServerService) DecryptCredential(server *model.Server) (string, error) {
	return pkg.Decrypt(server.Credential, s.encryptKey)
}

// ResolveServer fetches a server by ID and decrypts its credential in one call.
func (s *ServerService) ResolveServer(serverID uint) (*model.Server, string, error) {
	server, err := s.serverRepo.FindByID(serverID)
	if err != nil {
		return nil, "", errors.New("server not found")
	}
	cred, err := pkg.Decrypt(server.Credential, s.encryptKey)
	if err != nil {
		return nil, "", errors.New("failed to decrypt credential")
	}
	return server, cred, nil
}

func (s *ServerService) toResponse(server *model.Server) *dto.ServerResponse {
	return &dto.ServerResponse{
		ID:          server.ID,
		Host:        server.Host,
		Hostname:    server.Hostname,
		Port:        server.Port,
		User:        server.User,
		AuthType:    server.AuthType,
		Description: server.Description,
		CreatedAt:   server.CreatedAt,
		UpdatedAt:   server.UpdatedAt,
	}
}
