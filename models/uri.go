package models

// Uri:
type Uri struct {
	Absolute       bool      `json:"absolute"`
	Authority      Authority `json:"authority"`
	Empty          bool      `json:"empty"`
	Fragment       string    `json:"fragment"`
	Path           Path      `json:"path"`
	RawQueryString string    `json:"rawQueryString"`
	Relative       bool      `json:"relative"`
	Scheme         string    `json:"scheme"`
}
