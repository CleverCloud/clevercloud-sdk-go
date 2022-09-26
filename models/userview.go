package models

// UserView:
type UserView struct {
	Address        string   `json:"address"`
	Admin          bool     `json:"admin"`
	Avatar         string   `json:"avatar"`
	CanPay         bool     `json:"canPay"`
	City           string   `json:"city"`
	Country        string   `json:"country"`
	CreationDate   int64    `json:"creationDate"`
	Email          string   `json:"email"`
	EmailValidated bool     `json:"emailValidated"`
	HasPassword    bool     `json:"hasPassword"`
	ID             string   `json:"id"`
	Lang           string   `json:"lang"`
	Name           string   `json:"name"`
	OauthApps      []string `json:"oauthApps"`
	Phone          string   `json:"phone"`
	PreferredMFA   string   `json:"preferredMFA"`
	Zipcode        string   `json:"zipcode"`
}
