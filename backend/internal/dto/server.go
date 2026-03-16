package dto

import "time"

type CreateServerRequest struct {
	Host        string `json:"host" binding:"required"`
	Hostname    string `json:"hostname" binding:"required"`
	Port        int    `json:"port"`
	User        string `json:"user" binding:"required"`
	AuthType    string `json:"auth_type" binding:"required,oneof=password key"`
	Credential  string `json:"credential" binding:"required"`
	Description string `json:"description"`
}

type UpdateServerRequest struct {
	Host        string `json:"host"`
	Hostname    string `json:"hostname"`
	Port        int    `json:"port"`
	User        string `json:"user"`
	AuthType    string `json:"auth_type" binding:"omitempty,oneof=password key"`
	Credential  string `json:"credential"`
	Description string `json:"description"`
}

type ServerResponse struct {
	ID          uint      `json:"id"`
	Host        string    `json:"host"`
	Hostname    string    `json:"hostname"`
	Port        int       `json:"port"`
	User        string    `json:"user"`
	AuthType    string    `json:"auth_type"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TestConnectionRequest struct {
	Hostname   string `json:"hostname" binding:"required"`
	Port       int    `json:"port"`
	User       string `json:"user" binding:"required"`
	AuthType   string `json:"auth_type" binding:"required,oneof=password key"`
	Credential string `json:"credential" binding:"required"`
}
