package models

// SshKeyView:
type SshKeyView struct {
	Fingerprint string `json:"fingerprint"`
	Key         string `json:"key"`
	Name        string `json:"name"`
}
