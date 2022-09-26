package models

// RemoteAddress:
type RemoteAddress struct {
	Address interface{} `json:"address"`
	Port    int32       `json:"port"`
	Unknown bool        `json:"unknown"`
}
