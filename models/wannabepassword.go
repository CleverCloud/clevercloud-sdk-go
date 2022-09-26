package models

// WannabePassword:
type WannabePassword struct {
	DropTokens  bool   `json:"dropTokens"`
	NewPassword string `json:"newPassword"`
	OldPassword string `json:"oldPassword"`
}
