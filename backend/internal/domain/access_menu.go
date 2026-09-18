package domain

type AccessMenu struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
	Path string `json:"path"`
	Icon string `json:"icon"`
}

type RoleAccessConfig struct {
	Role      string `json:"role"`
	MenuKey   string `json:"menu_key"`
	CanAccess bool   `json:"can_access"`
}
