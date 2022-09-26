package models

// WannabeAddonProviderAPI:
type WannabeAddonProviderAPI struct {
	Config_vars []string                   `json:"config_vars"`
	Password    string                     `json:"password"`
	Production  WannabeAddonProviderAPIUrl `json:"production"`
	Regions     []string                   `json:"regions"`
	Sso_salt    string                     `json:"sso_salt"`
	Test        WannabeAddonProviderAPIUrl `json:"test"`
}
