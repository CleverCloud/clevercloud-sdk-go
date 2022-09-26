package models

// OwnerView:
type OwnerView struct {
	Emails   []string `json:"emails"`
	ID       string   `json:"id"`
	IDentity string   `json:"identity"`
}
