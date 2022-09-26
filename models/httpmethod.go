package models

// HttpMethod:
type HttpMethod struct {
	EntityAccepted          bool                    `json:"entityAccepted"`
	IsIDempotent            bool                    `json:"isIdempotent"`
	IsSafe                  bool                    `json:"isSafe"`
	RequestEntityAcceptance RequestEntityAcceptance `json:"requestEntityAcceptance"`
	Value                   string                  `json:"value"`
}
