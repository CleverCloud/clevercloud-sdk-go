package models

// WannabeNamespace:
type WannabeNamespace struct {
	MaxPort   int64  `json:"maxPort"`
	MinPort   int64  `json:"minPort"`
	Namespace string `json:"namespace"`
}
