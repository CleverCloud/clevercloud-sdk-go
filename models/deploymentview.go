package models

// DeploymentView:
type DeploymentView struct {
	Action    string `json:"action"`
	Author    Author `json:"author"`
	Cause     string `json:"cause"`
	Commit    string `json:"commit"`
	Date      int64  `json:"date"`
	ID        int32  `json:"id"`
	Instances int32  `json:"instances"`
	State     string `json:"state"`
	Uuid      string `json:"uuid"`
}
