package models

// WannabeApplication:
type WannabeApplication struct {
	ApplianceID      string          `json:"applianceId"`
	Archived         bool            `json:"archived"`
	Branch           string          `json:"branch"`
	BuildFlavor      string          `json:"buildFlavor"`
	CancelOnPush     bool            `json:"cancelOnPush"`
	Deploy           string          `json:"deploy"`
	Description      string          `json:"description"`
	Favourite        bool            `json:"favourite"`
	ForceHttps       string          `json:"forceHttps"`
	Homogeneous      bool            `json:"homogeneous"`
	InstanceLifetime string          `json:"instanceLifetime"`
	InstanceType     string          `json:"instanceType"`
	InstanceVariant  string          `json:"instanceVariant"`
	InstanceVersion  string          `json:"instanceVersion"`
	MaxFlavor        string          `json:"maxFlavor"`
	MaxInstances     int32           `json:"maxInstances"`
	MinFlavor        string          `json:"minFlavor"`
	MinInstances     int32           `json:"minInstances"`
	Name             string          `json:"name"`
	OauthApp         WannabeOauthApp `json:"oauthApp"`
	OauthAppID       string          `json:"oauthAppId"`
	OauthService     string          `json:"oauthService"`
	SeparateBuild    bool            `json:"separateBuild"`
	Shutdownable     bool            `json:"shutdownable"`
	StickySessions   bool            `json:"stickySessions"`
	Tags             []string        `json:"tags"`
	Zone             string          `json:"zone"`
}
