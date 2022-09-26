package models

// OrganisationMemberUserView:
type OrganisationMemberUserView struct {
	Avatar       string `json:"avatar"`
	Email        string `json:"email"`
	ID           string `json:"id"`
	Name         string `json:"name"`
	PreferredMFA string `json:"preferredMFA"`
}
