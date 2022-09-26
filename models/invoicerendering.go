package models

// InvoiceRendering:
type InvoiceRendering struct {
	Address            string                 `json:"address"`
	City               string                 `json:"city"`
	Company            string                 `json:"company"`
	Country            string                 `json:"country"`
	CountryCode        string                 `json:"countryCode"`
	CustomerCostCenter string                 `json:"customerCostCenter"`
	CustomerOrderID    string                 `json:"customerOrderId"`
	EmissionDate       string                 `json:"emissionDate"`
	ErrorCode          string                 `json:"errorCode"`
	ErrorLongMsg       string                 `json:"errorLongMsg"`
	ErrorShortMsg      string                 `json:"errorShortMsg"`
	ForID              string                 `json:"forId"`
	FromSubscription   bool                   `json:"fromSubscription"`
	ID                 string                 `json:"id"`
	Lines              []InvoiceLineRendering `json:"lines"`
	Name               string                 `json:"name"`
	Number             string                 `json:"number"`
	OriginalTotal      float64                `json:"originalTotal"`
	PayWhen            string                 `json:"payWhen"`
	PayingUser         OrganisationMemberView `json:"payingUser"`
	PaymentDate        string                 `json:"paymentDate"`
	Status             string                 `json:"status"`
	Target             string                 `json:"target"`
	TargetID           string                 `json:"targetId"`
	Token              string                 `json:"token"`
	TotalHT            float64                `json:"totalHT"`
	TotalTTC           float64                `json:"totalTTC"`
	TotalVAT           float64                `json:"totalVAT"`
	Type               string                 `json:"type"`
	VatNumber          string                 `json:"vatNumber"`
	VatRate            float64                `json:"vatRate"`
	ZipCode            string                 `json:"zipCode"`
}
