package models

// DropCountView:
type DropCountView struct {
	Count     float64       `json:"count"`
	DropPrice DropPriceView `json:"dropPrice"`
	OwnerID   string        `json:"ownerId"`
}
