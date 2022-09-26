package models

// RecurrentPaymentView:
type RecurrentPaymentView struct {
	Currency        string    `json:"currency"`
	MonthlyAmount   float64   `json:"monthlyAmount"`
	Target          OwnerView `json:"target"`
	Threshold       float64   `json:"threshold"`
	ThresholdAmount float64   `json:"thresholdAmount"`
	Token           string    `json:"token"`
	User            UserView  `json:"user"`
}
