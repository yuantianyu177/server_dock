package dto

import "time"

type SubmitApplicationRequest struct {
	ApplicantName  string `json:"applicant_name" binding:"required"`
	ApplicantEmail string `json:"applicant_email" binding:"required,email"`
	ServerID       uint   `json:"server_id" binding:"required"`
	ImageID        uint   `json:"image_id" binding:"required"`
}

type ApplicationActionRequest struct {
	Action string `json:"action" binding:"required,oneof=approve reject ignore"`
}

type EmailApplicationActionRequest struct {
	Token string `json:"token" binding:"required"`
}

type EmailApplicationActionResponse struct {
	Status  string `json:"status"`
	Title   string `json:"title"`
	Message string `json:"message"`
}

type ApplicationResponse struct {
	ID             uint                     `json:"id"`
	ApplicantName  string                   `json:"applicant_name"`
	ApplicantEmail string                   `json:"applicant_email"`
	ServerID       uint                     `json:"server_id"`
	ServerHost     string                   `json:"server_host"`
	ImageID        uint                     `json:"image_id"`
	ImageName      string                   `json:"image_name"`
	Status         string                   `json:"status"`
	CreatedAt      time.Time                `json:"created_at"`
	UpdatedAt      time.Time                `json:"updated_at"`
	ConnectionInfo *ContainerConnectionInfo `json:"connection_info,omitempty" gorm:"-"`
}

type ContainerConnectionInfo struct {
	Server     string `json:"server"`
	User       string `json:"user"`
	Password   string `json:"password"`
	SSHPort    int    `json:"ssh_port"`
	ExtraPorts string `json:"extra_ports"`
	SSHCommand string `json:"ssh_command"`
}

type PublicServerInfo struct {
	ID                uint   `json:"id"`
	Host              string `json:"host"`
	Description       string `json:"description"`
	RunningContainers int    `json:"running_containers"`
	TotalContainers   int    `json:"total_containers"`
	LoadAvailable     bool   `json:"load_available"`
}

type PublicImageInfo struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	ImageAddress string `json:"image_address"`
}
