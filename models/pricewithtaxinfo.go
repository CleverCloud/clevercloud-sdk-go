package models

// PriceWithTaxInfo:
type PriceWithTaxInfo struct {
	InvoicedOwner string `json:"invoicedOwner"`
	Price         int64  `json:"price"`
	TaxFreePrice  int64  `json:"taxFreePrice"`
	VatRatio      int64  `json:"vatRatio"`
}
