package models

// OAuthTransactionView:
type OAuthTransactionView struct {
	RedirectUri   string `json:"redirectUri"`
	TransactionID string `json:"transactionId"`
}
