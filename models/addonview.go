package models

// AddonView:
type AddonView struct {
	ConfigKeys   []string              `json:"configKeys"`
	CreationDate int64                 `json:"creationDate"`
	ID           string                `json:"id"`
	Name         string                `json:"name"`
	Plan         AddonPlanView         `json:"plan"`
	Provider     AddonProviderInfoView `json:"provider"`
	RealID       string                `json:"realId"`
	Region       string                `json:"region"`
}
