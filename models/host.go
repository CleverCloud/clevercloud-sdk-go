package models

// Host:
type Host struct {
	Empty         bool        `json:"empty"`
	InetAddresses interface{} `json:"inetAddresses"`
	Ipv4          bool        `json:"ipv4"`
	Ipv6          bool        `json:"ipv6"`
	NamedHost     bool        `json:"namedHost"`
}
