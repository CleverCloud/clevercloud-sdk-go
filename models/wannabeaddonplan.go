package models

// WannabeAddonPlan:
type WannabeAddonPlan struct {
	Features []AddonFeatureInstanceView `json:"features"`
	Name     string                     `json:"name"`
	Price    float64                    `json:"price"`
	Slug     string                     `json:"slug"`
}
