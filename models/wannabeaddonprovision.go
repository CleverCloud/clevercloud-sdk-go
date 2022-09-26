package models

// WannabeAddonProvision:
type WannabeAddonProvision struct {
	LinkedApp string `json:"linkedApp"`
	// Name of the future add-on, for display.
	Name string `json:"name"`
	// Options to add to the provision call.
	Options       interface{}     `json:"options"`
	PaymentIntent SetupIntentView `json:"paymentIntent"`
	// Payment method type
	PaymentMethodType string `json:"paymentMethodType"`
	// Id of the price plan
	Plan string `json:"plan"`
	// Id of the add-on provider
	ProviderID string `json:"providerId"`
	// Region to provision the add-on in.
	Region string `json:"region"`
	// Id of the SEPA debit source
	SepaSourceID string `json:"sepaSourceId"`
	// Version of the add-on
	Version string `json:"version"`
}
