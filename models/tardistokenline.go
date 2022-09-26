package models

// TardisTokenLine:
type TardisTokenLine struct {
	Addon_id            string          `json:"addon_id"`
	Creation_date       string          `json:"creation_date"`
	Expiration_date     string          `json:"expiration_date"`
	Name                string          `json:"name"`
	Properties          string          `json:"properties"`
	Senx_application_id string          `json:"senx_application_id"`
	Token               string          `json:"token"`
	TokenType           TardisTokenType `json:"tokenType"`
	Token_id            string          `json:"token_id"`
	Ttl                 string          `json:"ttl"`
}
