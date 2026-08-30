package service

import (
	"errors"

	"serverdock/internal/dto"
	"serverdock/internal/model"
	"serverdock/internal/pkg"

	"gorm.io/gorm"
)

type ServerService struct {
	db             *gorm.DB
	testConnection SSHTestFunc
	runCommand     SSHRunFunc
	encryptKey     string
}

func NewServerService(db *gorm.DB, testConnection SSHTestFunc, runCommand SSHRunFunc, encryptKey string) *ServerService {
	return &ServerService{db: db, testConnection: testConnection, runCommand: runCommand, encryptKey: encryptKey}
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
		Host: req.Host, Hostname: req.Hostname, Port: port, User: req.User,
		AuthType: req.AuthType, Credential: encrypted, Description: req.Description,
	}
	if err := s.db.Create(server).Error; err != nil {
		return nil, err
	}
	return serverResponse(server), nil
}

func (s *ServerService) GetByID(id uint) (*dto.ServerResponse, error) {
	server, err := s.find(id)
	if err != nil {
		return nil, errors.New("server not found")
	}
	return serverResponse(server), nil
}

func (s *ServerService) List() ([]dto.ServerResponse, error) {
	var servers []model.Server
	if err := s.db.Order("id desc").Find(&servers).Error; err != nil {
		return nil, err
	}
	responses := make([]dto.ServerResponse, len(servers))
	for i := range servers {
		responses[i] = *serverResponse(&servers[i])
	}
	return responses, nil
}

func (s *ServerService) Update(id uint, req *dto.UpdateServerRequest) (*dto.ServerResponse, error) {
	server, err := s.find(id)
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
		server.Credential, err = pkg.Encrypt(req.Credential, s.encryptKey)
		if err != nil {
			return nil, errors.New("failed to encrypt credential")
		}
	}
	server.Description = req.Description

	if err := s.db.Save(server).Error; err != nil {
		return nil, err
	}
	return serverResponse(server), nil
}

func (s *ServerService) Delete(id uint) error {
	result := s.db.Delete(&model.Server{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("server not found")
	}
	return nil
}

func (s *ServerService) TestConnectionDirect(hostname string, port int, user, authType, credential string) error {
	if port <= 0 {
		port = 22
	}
	return s.testConnection(hostname, port, user, authType, credential)
}

func (s *ServerService) TestConnection(id uint) error {
	server, credential, err := s.ResolveServer(id)
	if err != nil {
		return err
	}
	return s.testConnection(server.Hostname, server.Port, server.User, server.AuthType, credential)
}

func (s *ServerService) ResolveServer(id uint) (*model.Server, string, error) {
	server, err := s.find(id)
	if err != nil {
		return nil, "", errors.New("server not found")
	}
	credential, err := pkg.Decrypt(server.Credential, s.encryptKey)
	if err != nil {
		return nil, "", errors.New("failed to decrypt credential")
	}
	return server, credential, nil
}

func (s *ServerService) ExecuteCommand(serverID uint, command string) (string, error) {
	server, credential, err := s.ResolveServer(serverID)
	if err != nil {
		return "", err
	}
	return s.runCommand(server.Hostname, server.Port, server.User, server.AuthType, credential, command)
}

func (s *ServerService) find(id uint) (*model.Server, error) {
	var server model.Server
	return &server, s.db.First(&server, id).Error
}

func serverResponse(server *model.Server) *dto.ServerResponse {
	return &dto.ServerResponse{
		ID: server.ID, Host: server.Host, Hostname: server.Hostname, Port: server.Port,
		User: server.User, AuthType: server.AuthType, Description: server.Description,
		CreatedAt: server.CreatedAt, UpdatedAt: server.UpdatedAt,
	}
}
