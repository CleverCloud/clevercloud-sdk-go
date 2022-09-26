package models

// AddonApplicationSummary:
type AddonApplicationSummary struct {
	Callback_url string `json:"callback_url"`
	Heroku_id    string `json:"heroku_id"`
	Owner_id     string `json:"owner_id"`
	Plan         string `json:"plan"`
	Provider_id  string `json:"provider_id"`
}
