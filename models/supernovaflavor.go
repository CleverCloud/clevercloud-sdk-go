package models

// SuperNovaFlavor:
type SuperNovaFlavor struct {
	Cpus  int32   `json:"cpus"`
	Mem   int32   `json:"mem"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}
