package dto

import "time"

type SubmitApplicationRequest struct {
	ApplicantName  string `json:"applicant_name" binding:"required"`
	ApplicantEmail string `json:"applicant_email" binding:"required,email"`
	ServerID       uint   `json:"server_id" binding:"required"`
	ImageID        uint   `json:"image_id" binding:"required"`
}

type ApplicationActionRequest struct {
	Action     string `json:"action" binding:"required,oneof=approve reject"`
	AdminNotes string `json:"admin_notes"`
}

type ApplicationResponse struct {
	ID             uint      `json:"id"`
	ApplicantName  string    `json:"applicant_name"`
	ApplicantEmail string    `json:"applicant_email"`
	ServerID       uint      `json:"server_id"`
	ServerHost     string    `json:"server_host"`
	ImageID        uint      `json:"image_id"`
	ImageName      string    `json:"image_name"`
	Status         string    `json:"status"`
	AdminNotes     string    `json:"admin_notes"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type PublicServerInfo struct {
	ID             uint   `json:"id"`
	Host           string `json:"host"`
	Description    string `json:"description"`
	ContainerCount int    `json:"container_count"`
}

type PublicImageInfo struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	ImageAddress string `json:"image_address"`
}
