package model

type Access struct {
	App          string `json:"app" mapstructure:"app"`
	Admin        bool   `json:"admin" mapstructure:"admin"`
	PermissionID string `json:"permission_set_id" mapstructure:"permission_set_id"`
}
