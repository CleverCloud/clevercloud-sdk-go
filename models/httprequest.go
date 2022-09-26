package models

// HttpRequest:
type HttpRequest struct {
	Attributes interface{}   `json:"attributes"`
	Entity     RequestEntity `json:"entity"`
	Headers    interface{}   `json:"headers"`
	Method     HttpMethod    `json:"method"`
	Protocol   HttpProtocol  `json:"protocol"`
	Request    bool          `json:"request"`
	Response   bool          `json:"response"`
	Uri        Uri           `json:"uri"`
}
