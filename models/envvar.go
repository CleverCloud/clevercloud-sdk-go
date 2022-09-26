package models

// EnvVar:
type EnvVar struct {
	// Ex: PORT
	Name string `json:"name"`
	// Ex: 8080
	Value string `json:"value"`
}
