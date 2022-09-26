package models

// OrganisationMemberView:
type OrganisationMemberView struct {
	Job    string                     `json:"job"`
	Member OrganisationMemberUserView `json:"member"`
	Role   string                     `json:"role"`
}
