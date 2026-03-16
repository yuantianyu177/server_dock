package dto

type UpdateConfigRequest struct {
	Value string `json:"value"`
}

type ConfigItem struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
