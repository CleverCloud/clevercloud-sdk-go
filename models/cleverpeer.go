package models

// CleverPeer:
type CleverPeer struct {
	Endpoint      WireguardEndpoint  `json:"endpoint"`
	Hostname      PeerHostName       `json:"hostname"`
	Hv            HypervisorID       `json:"hv"`
	ID            PeerID             `json:"id"`
	Label         PeerLabel          `json:"label"`
	Parent_event  PeerParentEvent    `json:"parent_event"`
	Parent_member MemberID           `json:"parent_member"`
	Public_key    WireGuardPublicKey `json:"public_key"`
}
