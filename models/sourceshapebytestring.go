package models

// SourceShapeByteString:
type SourceShapeByteString struct {
	Inlets  []InletObject    `json:"inlets"`
	Out     OutletByteString `json:"out"`
	Outlets []OutletObject   `json:"outlets"`
}
