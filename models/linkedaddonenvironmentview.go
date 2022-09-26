package models

// LinkedAddonEnvironmentView:
type LinkedAddonEnvironmentView struct {
	Addon_id    string                 `json:"addon_id"`
	Addon_name  string                 `json:"addon_name"`
	Env         []AddonEnvironmentView `json:"env"`
	Provider_id string                 `json:"provider_id"`
}
