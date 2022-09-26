package models

// SetupIntentView:
// PaymentIntent id from Stripe
type SetupIntentView struct {
	ClientSecret string `json:"clientSecret"`
	Customer     string `json:"customer"`
	ID           string `json:"id"`
	OwnerID      string `json:"ownerId"`
}
