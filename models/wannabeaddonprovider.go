package models

// WannabeAddonProvider:
type WannabeAddonProvider struct {
	Api  WannabeAddonProviderAPI `json:"api"`
	ID   string                  `json:"id"`
	Name string                  `json:"name"`
}
