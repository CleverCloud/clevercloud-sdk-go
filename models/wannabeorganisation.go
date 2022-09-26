package models

// WannabeOrganisation:
type WannabeOrganisation struct {
	VAT              string `json:"VAT"`
	Address          string `json:"address"`
	BillingEmail     string `json:"billingEmail"`
	City             string `json:"city"`
	Company          string `json:"company"`
	Country          string `json:"country"`
	CustomerFullName string `json:"customerFullName"`
	Description      string `json:"description"`
	Name             string `json:"name"`
	Zipcode          string `json:"zipcode"`
}
