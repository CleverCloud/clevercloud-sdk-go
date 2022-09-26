package models

// Authority:
type Authority struct {
	Empty    bool   `json:"empty"`
	Host     Host   `json:"host"`
	Port     int32  `json:"port"`
	Userinfo string `json:"userinfo"`
}
