package models

// ExternalPeer:
type ExternalPeer struct {
	Endpoint      WireguardEndpoint  `json:"endpoint"`
	Hostname      PeerHostName       `json:"hostname"`
	ID            PeerID             `json:"id"`
	Label         PeerLabel          `json:"label"`
	Parent_event  PeerParentEvent    `json:"parent_event"`
	Parent_member MemberID           `json:"parent_member"`
	Public_key    WireGuardPublicKey `json:"public_key"`
}
