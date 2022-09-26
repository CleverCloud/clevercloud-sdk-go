package models

// OAuth1AccessTokenView:
type OAuth1AccessTokenView struct {
	Consumer        OAuth1ConsumerView `json:"consumer"`
	CreationDate    int64              `json:"creationDate"`
	LastUtilisation int64              `json:"lastUtilisation"`
	Rights          OAuthRightsView    `json:"rights"`
	Token           string             `json:"token"`
}
