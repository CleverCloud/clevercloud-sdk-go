package models

// WannabeEnvVar:
type WannabeEnvVar struct {
	// Ex: DEBUG
	Name string `json:"name"`
	// Ex: TRUE
	Value string `json:"value"`
}
