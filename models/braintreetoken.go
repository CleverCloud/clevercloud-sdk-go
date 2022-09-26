package models

// BraintreeToken:
type BraintreeToken struct {
	MerchantID string `json:"merchantId"`
	Token      string `json:"token"`
}
