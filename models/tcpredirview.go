package models

// TcpRedirView:
type TcpRedirView struct {
	Namespace string `json:"namespace"`
	Port      int64  `json:"port"`
}
