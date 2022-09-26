package models

// UserSummary:
type UserSummary struct {
	Addons       []AddonSummary          `json:"addons"`
	Admin        bool                    `json:"admin"`
	Applications []ApplicationSummary    `json:"applications"`
	Avatar       string                  `json:"avatar"`
	CanSEPA      bool                    `json:"canSEPA"`
	Consumers    []OAuth1ConsumerSummary `json:"consumers"`
	Email        string                  `json:"email"`
	ID           string                  `json:"id"`
	Lang         string                  `json:"lang"`
	Name         string                  `json:"name"`
}
