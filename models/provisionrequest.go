package models

// ProvisionRequest:
type ProvisionRequest struct {
	Callback_url   string      `json:"callback_url"`
	Heroku_id      string      `json:"heroku_id"`
	Logplex_token  string      `json:"logplex_token"`
	Logs_collector string      `json:"logs_collector"`
	Options        interface{} `json:"options"`
	Owner_id       string      `json:"owner_id"`
	Plan           string      `json:"plan"`
	Region         string      `json:"region"`
	User_id        string      `json:"user_id"`
	Version        string      `json:"version"`
}
