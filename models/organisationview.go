package models

// OrganisationView:
type OrganisationView struct {
	VAT              string `json:"VAT"`
	Address          string `json:"address"`
	Avatar           string `json:"avatar"`
	BillingEmail     string `json:"billingEmail"`
	CanPay           bool   `json:"canPay"`
	CanSEPA          bool   `json:"canSEPA"`
	City             string `json:"city"`
	CleverEnterprise bool   `json:"cleverEnterprise"`
	Company          string `json:"company"`
	Country          string `json:"country"`
	CustomerFullName string `json:"customerFullName"`
	Description      string `json:"description"`
	EmergencyNumber  string `json:"emergencyNumber"`
	ID               string `json:"id"`
	IsTrusted        bool   `json:"isTrusted"`
	Name             string `json:"name"`
	VatState         string `json:"vatState"`
	Zipcode          string `json:"zipcode"`
}
