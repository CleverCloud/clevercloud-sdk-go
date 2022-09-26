package models

// AddonConfigProviderView:
type AddonConfigProviderView struct {
	Creation_date string                  `json:"creation_date"`
	Env           []EnvVar                `json:"env"`
	ID            ConfigurationProviderID `json:"id"`
	Owner_id      OwnerID                 `json:"owner_id"`
}
