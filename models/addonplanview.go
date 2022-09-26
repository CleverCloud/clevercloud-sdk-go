package models

// AddonPlanView:
type AddonPlanView struct {
	Features []AddonFeatureInstanceView `json:"features"`
	ID       string                     `json:"id"`
	Name     string                     `json:"name"`
	Price    float64                    `json:"price"`
	Price_id string                     `json:"price_id"`
	Slug     string                     `json:"slug"`
	Zones    []string                   `json:"zones"`
}
