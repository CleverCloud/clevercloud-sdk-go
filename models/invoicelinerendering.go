package models

// InvoiceLineRendering:
type InvoiceLineRendering struct {
	VAT            float64 `json:"VAT"`
	Description    string  `json:"description"`
	Pack           string  `json:"pack"`
	Price          float64 `json:"price"`
	SubDescription string  `json:"subDescription"`
}
