package models

// WannaBeApiTokenRoute:
type WannaBeApiTokenRoute struct {
	Ephemeral  interface{}     `json:"ephemeral"`
	Name       string          `json:"name"`
	Properties MapStringString `json:"properties"`
	TokenType  TardisTokenType `json:"tokenType"`
	Ttl        string          `json:"ttl"`
}
