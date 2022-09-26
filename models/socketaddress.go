package models

// SocketAddress: SocketAddress
type SocketAddress struct {
	// Ex: 192.168.1.1
	Ip string `json:"ip"`
	// Ex: %!s(float64=8080)
	Port int64 `json:"port"`
}
