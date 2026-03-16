package dto

import "time"

type CreateImageRequest struct {
	ServerID     uint   `json:"server_id" binding:"required"`
	ImageID      string `json:"image_id" binding:"required"`
	Name         string `json:"name" binding:"required"`
	ImageAddress string `json:"image_address" binding:"required"`
}

type UpdateImageRequest struct {
	Name         string `json:"name"`
	ImageAddress string `json:"image_address"`
}

type ImageResponse struct {
	ID           uint      `json:"id"`
	ServerID     uint      `json:"server_id"`
	ImageID      string    `json:"image_id"`
	Name         string    `json:"name"`
	ImageAddress string    `json:"image_address"`
	CreatedAt    time.Time `json:"created_at"`
}

type RemoteImage struct {
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
	ImageID    string `json:"image_id"`
	Size       string `json:"size"`
	Created    string `json:"created"`
}

type PullImageRequest struct {
	Image string `json:"image" binding:"required"`
}
