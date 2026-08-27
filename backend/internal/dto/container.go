package dto

type CreateContainerRequest struct {
	Name      string `json:"name" binding:"required"`
	Image     string `json:"image" binding:"required"`
	ExtraArgs string `json:"extra_args"`
}

type ContainerActionRequest struct {
	Action string `json:"action" binding:"required,oneof=start stop restart delete"`
}

type CreateVolumeRequest struct {
	Name string `json:"name" binding:"required"`
}
