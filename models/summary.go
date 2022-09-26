package models

// Summary:
type Summary struct {
	Organisations []OrganisationSummary `json:"organisations"`
	User          UserSummary           `json:"user"`
}
