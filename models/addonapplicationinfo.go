package models

// AddonApplicationInfo:
type AddonApplicationInfo struct {
	Callback_url string      `json:"callback_url"`
	Config       interface{} `json:"config"`
	Domains      []string    `json:"domains"`
	ID           string      `json:"id"`
	Name         string      `json:"name"`
	Owner_email  string      `json:"owner_email"`
	Owner_emails []string    `json:"owner_emails"`
	Owner_id     string      `json:"owner_id"`
	Owner_name   string      `json:"owner_name"`
	Region       string      `json:"region"`
}
