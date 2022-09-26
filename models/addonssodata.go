package models

// AddonSSOData:
type AddonSSOData struct {
	Email     string `json:"email"`
	ID        string `json:"id"`
	NavData   string `json:"nav-data"`
	Timestamp int64  `json:"timestamp"`
	Token     string `json:"token"`
	Url       string `json:"url"`
}
