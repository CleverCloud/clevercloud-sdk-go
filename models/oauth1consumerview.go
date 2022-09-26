package models

// OAuth1ConsumerView:
type OAuth1ConsumerView struct {
	BaseUrl     string          `json:"baseUrl"`
	Description string          `json:"description"`
	Key         string          `json:"key"`
	Name        string          `json:"name"`
	Picture     string          `json:"picture"`
	Rights      OAuthRightsView `json:"rights"`
	Url         string          `json:"url"`
}
