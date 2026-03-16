package dto

type CreateContainerRequest struct {
	Name      string `json:"name" binding:"required"`
	Image     string `json:"image" binding:"required"`
	ExtraArgs string `json:"extra_args"`
}

type ContainerActionRequest struct {
	Action string `json:"action" binding:"required,oneof=start stop restart delete"`
}

type ContainerLogsRequest struct {
	Tail int `form:"tail"`
}

type ExecCommandRequest struct {
	Command string `json:"command" binding:"required"`
}

type CreateVolumeRequest struct {
	Name string `json:"name" binding:"required"`
}
