package models

// ApplicationView:
type ApplicationView struct {
	Appliance      string             `json:"appliance"`
	Archived       bool               `json:"archived"`
	Branch         string             `json:"branch"`
	BuildFlavor    FlavorView         `json:"buildFlavor"`
	CancelOnPush   bool               `json:"cancelOnPush"`
	CommitID       string             `json:"commitId"`
	CreationDate   int64              `json:"creationDate"`
	DeployUrl      string             `json:"deployUrl"`
	Deployment     DeploymentInfoView `json:"deployment"`
	Description    string             `json:"description"`
	Favourite      bool               `json:"favourite"`
	ForceHttps     string             `json:"forceHttps"`
	Homogeneous    bool               `json:"homogeneous"`
	ID             string             `json:"id"`
	Instance       InstanceView       `json:"instance"`
	Last_deploy    int32              `json:"last_deploy"`
	Name           string             `json:"name"`
	OwnerID        string             `json:"ownerId"`
	SeparateBuild  bool               `json:"separateBuild"`
	State          string             `json:"state"`
	StickySessions bool               `json:"stickySessions"`
	Vhosts         []VhostView        `json:"vhosts"`
	WebhookSecret  string             `json:"webhookSecret"`
	WebhookUrl     string             `json:"webhookUrl"`
	Zone           string             `json:"zone"`
}
