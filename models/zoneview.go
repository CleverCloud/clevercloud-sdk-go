package models

// ZoneView:
type ZoneView struct {
	CorrespondingRegion string `json:"correspondingRegion"`
	Internal            bool   `json:"internal"`
	Name                string `json:"name"`
}
