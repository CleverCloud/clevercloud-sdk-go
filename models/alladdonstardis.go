package models

// AllAddonsTardis:
type AllAddonsTardis struct {
	Addons  []AddonTardis `json:"addons"`
	OwnerID string        `json:"ownerID"`
}
