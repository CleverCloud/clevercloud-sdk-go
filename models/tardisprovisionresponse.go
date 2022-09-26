package models

// TardisProvisionResponse:
type TardisProvisionResponse struct {
	Config  interface{} `json:"config"`
	ID      TardisID    `json:"id"`
	Message string      `json:"message"`
}
