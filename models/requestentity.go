package models

// RequestEntity:
type RequestEntity struct {
	Chunked             bool        `json:"chunked"`
	CloseDelimited      bool        `json:"closeDelimited"`
	ContentLengthOption interface{} `json:"contentLengthOption"`
	ContentType         ContentType `json:"contentType"`
	DataBytes           interface{} `json:"dataBytes"`
	Default             bool        `json:"default"`
	IndefiniteLength    bool        `json:"indefiniteLength"`
	KnownEmpty          bool        `json:"knownEmpty"`
	Strict              bool        `json:"strict"`
}
