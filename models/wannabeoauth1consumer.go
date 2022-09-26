package models

// WannabeOAuth1Consumer:
type WannabeOAuth1Consumer struct {
	BaseUrl     string      `json:"baseUrl"`
	Description string      `json:"description"`
	Name        string      `json:"name"`
	Picture     string      `json:"picture"`
	Rights      interface{} `json:"rights"`
	Url         string      `json:"url"`
}
