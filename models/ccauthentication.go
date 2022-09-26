package models

// CCAuthentication:
type CCAuthentication struct {
	Authorization string     `json:"authorization"`
	Ip_request    string     `json:"ip_request"`
	Mac           string     `json:"mac"`
	Method        HttpMethod `json:"method"`
	Nonce         string     `json:"nonce"`
	Payload       string     `json:"payload"`
	Uri           Uri        `json:"uri"`
}
