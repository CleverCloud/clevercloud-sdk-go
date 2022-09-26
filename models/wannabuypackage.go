package models

// WannaBuyPackage:
type WannaBuyPackage struct {
	Coupon       string  `json:"coupon"`
	Currency     string  `json:"currency"`
	DropQuantity float64 `json:"dropQuantity"`
	PackageID    int64   `json:"packageId"`
}
