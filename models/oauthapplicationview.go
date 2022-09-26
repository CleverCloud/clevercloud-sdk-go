package models

// OAuthApplicationView:
type OAuthApplicationView struct {
	DefaultBranch string `json:"defaultBranch"`
	Description   string `json:"description"`
	GitUrl        string `json:"gitUrl"`
	ID            string `json:"id"`
	Name          string `json:"name"`
	Owner         string `json:"owner"`
	Priv          bool   `json:"priv"`
}
