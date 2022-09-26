package models

// OVDInvalidResponseBody:
type OVDInvalidResponseBody struct {
	Api_request_id string          `json:"api_request_id"`
	Code           string          `json:"code"`
	Context        OVDErrorContext `json:"context"`
	Error          string          `json:"error"`
}
