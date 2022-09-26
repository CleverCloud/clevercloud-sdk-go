package models

// NetworkGroup: Network Group
// Details about a Network Group.
type NetworkGroup struct {
	Description       NetworkGroupDescription            `json:"description"`
	ID                NetworkGroupID                     `json:"id"`
	Label             NetworkGroupLabel                  `json:"label"`
	Last_allocated_ip NetworkGroupLastAllocatedIpAddress `json:"last_allocated_ip"`
	Members           []Member                           `json:"members"`
	Network_ip        NetworkGroupIpAddress              `json:"network_ip"`
	Owner_id          OwnerID                            `json:"owner_id"`
	Peers             []Peer                             `json:"peers"`
	Tags              []string                           `json:"tags"`
	Version           NetworkGroupVersionNumber          `json:"version"`
}
