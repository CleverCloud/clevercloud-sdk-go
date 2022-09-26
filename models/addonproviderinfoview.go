package models

// AddonProviderInfoView:
type AddonProviderInfoView struct {
	AnalyticsID    string   `json:"analyticsId"`
	CanUpgrade     bool     `json:"canUpgrade"`
	GooglePlusName string   `json:"googlePlusName"`
	ID             string   `json:"id"`
	LogoUrl        string   `json:"logoUrl"`
	LongDesc       string   `json:"longDesc"`
	Name           string   `json:"name"`
	OpenInNewTab   bool     `json:"openInNewTab"`
	Regions        []string `json:"regions"`
	ShortDesc      string   `json:"shortDesc"`
	Status         string   `json:"status"`
	SupportEmail   string   `json:"supportEmail"`
	TwitterName    string   `json:"twitterName"`
	Website        string   `json:"website"`
}
