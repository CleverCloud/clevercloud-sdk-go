package models

// PaymentInfoView:
type PaymentInfoView struct {
	Address   string      `json:"address"`
	City      string      `json:"city"`
	Company   string      `json:"company"`
	Country   string      `json:"country"`
	Errors    interface{} `json:"errors"`
	FromOrga  bool        `json:"fromOrga"`
	Name      string      `json:"name"`
	VatNumber string      `json:"vatNumber"`
	ZipCode   string      `json:"zipCode"`
}
