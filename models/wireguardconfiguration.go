package models

// WireGuardConfiguration: WireGuard® Configuration
// WireGuard® configuration for a specific Network Group peer.
type WireGuardConfiguration struct {
	Base64  string                    `json:"base64"`
	Conf    WireGuardConfFile         `json:"conf"`
	Ng_id   NetworkGroupID            `json:"ng_id"`
	Peer_id PeerID                    `json:"peer_id"`
	Version NetworkGroupVersionNumber `json:"version"`
}
