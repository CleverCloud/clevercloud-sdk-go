package models

// WannabeInterAddonProvision:
type WannabeInterAddonProvision struct {
	AddonID        string      `json:"addonId"`
	Name           string      `json:"name"`
	Options        interface{} `json:"options"`
	OrganisationID string      `json:"organisationId"`
	Plan           string      `json:"plan"`
	ProviderID     string      `json:"providerId"`
	Region         string      `json:"region"`
	UserID         string      `json:"userId"`
}
