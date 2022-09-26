package models

// EndOfInvoiceError:
type EndOfInvoiceError struct {
	Message  string      `json:"message"`
	Messages interface{} `json:"messages"`
	Type     string      `json:"type"`
}
