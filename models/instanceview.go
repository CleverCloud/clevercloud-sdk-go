package models

// InstanceView:
type InstanceView struct {
	DefaultEnv          interface{}         `json:"defaultEnv"`
	Flavors             []FlavorView        `json:"flavors"`
	InstanceAndVersion  string              `json:"instanceAndVersion"`
	Lifetime            string              `json:"lifetime"`
	MaxAllowedInstances int32               `json:"maxAllowedInstances"`
	MaxFlavor           FlavorView          `json:"maxFlavor"`
	MaxInstances        int32               `json:"maxInstances"`
	MinFlavor           FlavorView          `json:"minFlavor"`
	MinInstances        int32               `json:"minInstances"`
	Type                string              `json:"type"`
	Variant             InstanceVariantView `json:"variant"`
	Version             string              `json:"version"`
}
