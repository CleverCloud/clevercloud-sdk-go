package models

// WannabeExternalPeer: Wannabe external peer
// Informations required to create an external Network Group peer.
type WannabeExternalPeer struct {
	Hostname      PeerHostName        `json:"hostname"`
	Ip            ServerPeerIpAddress `json:"ip"`
	Label         PeerLabel           `json:"label"`
	Parent_event  PeerParentEvent     `json:"parent_event"`
	Parent_member MemberID            `json:"parent_member"`
	// Peer Role
	// The role of this peer in the Network Group. Either client (`'client'`) or server (`'server'`).
	Peer_role  string             `json:"peer_role"`
	Port       ServerPeerPort     `json:"port"`
	Public_key WireGuardPublicKey `json:"public_key"`
}
