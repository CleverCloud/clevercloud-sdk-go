package models

// AvailableInstanceView:
type AvailableInstanceView struct {
	BuildFlavor   FlavorView          `json:"buildFlavor"`
	ComingSoon    bool                `json:"comingSoon"`
	DefaultFlavor FlavorView          `json:"defaultFlavor"`
	Deployments   []string            `json:"deployments"`
	Description   string              `json:"description"`
	Enabled       bool                `json:"enabled"`
	Flavors       []FlavorView        `json:"flavors"`
	MaxInstances  int32               `json:"maxInstances"`
	Name          string              `json:"name"`
	Tags          []string            `json:"tags"`
	Type          string              `json:"type"`
	Variant       InstanceVariantView `json:"variant"`
	Version       string              `json:"version"`
}
