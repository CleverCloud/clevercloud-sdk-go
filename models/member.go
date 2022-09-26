package models

// Member: Member
// Informations about a Network Group member.
type Member struct {
	Domain_name MemberName  `json:"domain_name"`
	ID          MemberID    `json:"id"`
	Label       MemberLabel `json:"label"`
	// Member Type
	// Type of a Network Group member.
	Type string `json:"type"`
}
