package models

// Attributes:
type Attributes struct {
	Async         bool        `json:"async"`
	AttributeList []Attribute `json:"attributeList"`
}
