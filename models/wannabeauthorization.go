package models

// WannabeAuthorization:
type WannabeAuthorization struct {
	Authorization string `json:"authorization"`
	Mac           string `json:"mac"`
	Nonce         string `json:"nonce"`
	Payload       string `json:"payload"`
	Uri           string `json:"uri"`
	Verb          string `json:"verb"`
}
