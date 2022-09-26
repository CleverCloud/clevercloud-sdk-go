package models

// ValueWithUnit:
type ValueWithUnit struct {
	Formatted string  `json:"formatted"`
	Unit      string  `json:"unit"`
	Value     float64 `json:"value"`
}
