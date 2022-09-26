package models

// OrganisationSummary:
type OrganisationSummary struct {
	Addons           []AddonSummary          `json:"addons"`
	Applications     []ApplicationSummary    `json:"applications"`
	Avatar           string                  `json:"avatar"`
	CanPay           bool                    `json:"canPay"`
	CanSEPA          bool                    `json:"canSEPA"`
	CleverEnterprise bool                    `json:"cleverEnterprise"`
	Consumers        []OAuth1ConsumerSummary `json:"consumers"`
	EmergencyNumber  string                  `json:"emergencyNumber"`
	ID               string                  `json:"id"`
	IsTrusted        bool                    `json:"isTrusted"`
	Name             string                  `json:"name"`
	Providers        []ProviderSummary       `json:"providers"`
	Role             string                  `json:"role"`
	VatState         string                  `json:"vatState"`
}
