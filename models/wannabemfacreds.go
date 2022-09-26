package models

// WannabeMFACreds:
type WannabeMFACreds struct {
	Code         string `json:"code"`
	RevokeTokens bool   `json:"revokeTokens"`
}
