package models

// PaymentMethodView:
type PaymentMethodView struct {
	BankCode       string `json:"bankCode"`
	BranchCode     string `json:"branchCode"`
	CardType       string `json:"cardType"`
	Country        string `json:"country"`
	Email          string `json:"email"`
	ExpirationDate string `json:"expirationDate"`
	Fingerprint    string `json:"fingerprint"`
	HolderName     string `json:"holderName"`
	ImageUrl       string `json:"imageUrl"`
	IsDefault      bool   `json:"isDefault"`
	IsExpired      bool   `json:"isExpired"`
	Number         string `json:"number"`
	OwnerID        string `json:"ownerId"`
	Token          string `json:"token"`
	Type           string `json:"type"`
}
