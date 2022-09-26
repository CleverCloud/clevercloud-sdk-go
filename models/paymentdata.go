package models

// PaymentData:
type PaymentData struct {
	DeviceData string `json:"deviceData"`
	Token      string `json:"token"`
	Type       string `json:"type"`
}
