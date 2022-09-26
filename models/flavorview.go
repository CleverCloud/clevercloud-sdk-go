package models

// FlavorView:
type FlavorView struct {
	Available        bool          `json:"available"`
	Cpus             int32         `json:"cpus"`
	Disk             int32         `json:"disk"`
	Gpus             int32         `json:"gpus"`
	Machine_learning bool          `json:"machine_learning"`
	Mem              int32         `json:"mem"`
	Memory           ValueWithUnit `json:"memory"`
	Microservice     bool          `json:"microservice"`
	Name             string        `json:"name"`
	Nice             int32         `json:"nice"`
	Price            float64       `json:"price"`
	Price_id         string        `json:"price_id"`
}
