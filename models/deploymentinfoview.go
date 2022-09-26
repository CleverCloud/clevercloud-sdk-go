package models

// DeploymentInfoView:
type DeploymentInfoView struct {
	HttpUrl      string `json:"httpUrl"`
	RepoState    string `json:"repoState"`
	Shutdownable bool   `json:"shutdownable"`
	Type         string `json:"type"`
	Url          string `json:"url"`
}
