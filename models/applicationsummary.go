package models

// ApplicationSummary:
type ApplicationSummary struct {
	Archived        bool   `json:"archived"`
	Commit          string `json:"commit"`
	Homogeneous     bool   `json:"homogeneous"`
	ID              string `json:"id"`
	InstanceType    string `json:"instanceType"`
	InstanceVariant string `json:"instanceVariant"`
	Name            string `json:"name"`
	State           string `json:"state"`
	VariantLogoUrl  string `json:"variantLogoUrl"`
	VariantSlug     string `json:"variantSlug"`
}
