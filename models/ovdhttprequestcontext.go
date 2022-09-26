package models

// OVDHttpRequestContext:
type OVDHttpRequestContext struct {
	Authentication  CCAuthentication `json:"authentication"`
	Client_ip       RemoteAddress    `json:"client_ip"`
	Request         HttpRequest      `json:"request"`
	Request_id      string           `json:"request_id"`
	Sozu_request_id string           `json:"sozu_request_id"`
}
