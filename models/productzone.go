package models

// ProductZone: Product zone definition
type ProductZone struct {
	// Ex: Paris
	City string `json:"city"`
	// Ex: France
	Country string `json:"country"`
	// Ex: FR
	CountryCode string `json:"countryCode"`
	// Ex: GPUs-enabled zone
	DisplayName string `json:"displayName"`
	// Ex: 7a50c50a-81bd-40f6-8a57-0068217097cf
	ID string `json:"id"`
	// Ex: %!s(float64=48.8566)
	Lat float64 `json:"lat"`
	// Ex: %!s(float64=2.3522)
	Lon float64 `json:"lon"`
	// Ex: par
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}
