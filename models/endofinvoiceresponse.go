package models

// EndOfInvoiceResponse:
type EndOfInvoiceResponse struct {
	BtcPrice       string            `json:"btcPrice"`
	Currency       string            `json:"currency"`
	CurrentTime    int64             `json:"currentTime"`
	Error          EndOfInvoiceError `json:"error"`
	ExpirationTime int64             `json:"expirationTime"`
	ID             string            `json:"id"`
	InvoiceTime    int64             `json:"invoiceTime"`
	PosData        string            `json:"posData"`
	Price          float64           `json:"price"`
	Status         string            `json:"status"`
	Url            string            `json:"url"`
}
