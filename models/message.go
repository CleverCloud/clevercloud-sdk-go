package models

// Message:
type Message struct {
	ID      int32  `json:"id"`
	Message string `json:"message"`
	Type    string `json:"type"`
}
