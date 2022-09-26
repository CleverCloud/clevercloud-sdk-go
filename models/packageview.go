package models

// PackageView:
type PackageView struct {
	DropPrice     DropPriceView `json:"dropPrice"`
	DropsQuantity float64       `json:"dropsQuantity"`
	ID            int64         `json:"id"`
	OriginalPrice float64       `json:"originalPrice"`
	Price         float64       `json:"price"`
	Title         string        `json:"title"`
}
