package models

// SuperNovaInstanceView:
type SuperNovaInstanceView struct {
	AppID          string          `json:"appId"`
	AppPort        int32           `json:"appPort"`
	Commit         string          `json:"commit"`
	CreationDate   int64           `json:"creationDate"`
	DeployID       string          `json:"deployId"`
	DeployNumber   int32           `json:"deployNumber"`
	DisplayName    string          `json:"displayName"`
	Flavor         SuperNovaFlavor `json:"flavor"`
	ID             string          `json:"id"`
	InstanceNumber int32           `json:"instanceNumber"`
	Ip             string          `json:"ip"`
	State          string          `json:"state"`
}
