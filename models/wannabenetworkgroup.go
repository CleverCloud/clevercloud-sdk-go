package models

// WannabeNetworkGroup: Wannabe Network Group
// Informations required to create an external Network Group.
type WannabeNetworkGroup struct {
	Description NetworkGroupDescription `json:"description"`
	ID          NetworkGroupID          `json:"id"`
	Label       NetworkGroupLabel       `json:"label"`
	Members     []WannabeMember         `json:"members"`
	Tags        []string                `json:"tags"`
}
