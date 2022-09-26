package models

// AddonSummary:
type AddonSummary struct {
	ID         string `json:"id"`
	LogoUrl    string `json:"logoUrl"`
	Name       string `json:"name"`
	ProviderID string `json:"providerId"`
	RealID     string `json:"realId"`
}
