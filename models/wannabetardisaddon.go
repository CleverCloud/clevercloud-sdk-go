package models

// WannabeTardisAddon:
type WannabeTardisAddon struct {
	AuthorID RefinedStringOrStartsWithStringOrEqualStringEqualString `json:"authorID"`
	OwnerID  string                                                  `json:"ownerID"`
	Plan     AddonTardisPlanIDentifier                               `json:"plan"`
	Ttl      string                                                  `json:"ttl"`
}
