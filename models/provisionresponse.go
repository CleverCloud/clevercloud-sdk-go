package models

// ProvisionResponse: Provision
type ProvisionResponse struct {
	Config  interface{} `json:"config"`
	ID      string      `json:"id"`
	Message string      `json:"message"`
}
