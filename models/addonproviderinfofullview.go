package models

// AddonProviderInfoFullView:
type AddonProviderInfoFullView struct {
	AnalyticsID    string             `json:"analyticsId"`
	CanUpgrade     bool               `json:"canUpgrade"`
	Features       []AddonFeatureView `json:"features"`
	GooglePlusName string             `json:"googlePlusName"`
	ID             string             `json:"id"`
	LogoUrl        string             `json:"logoUrl"`
	LongDesc       string             `json:"longDesc"`
	Name           string             `json:"name"`
	OpenInNewTab   bool               `json:"openInNewTab"`
	Plans          []AddonPlanView    `json:"plans"`
	Regions        []string           `json:"regions"`
	ShortDesc      string             `json:"shortDesc"`
	Status         string             `json:"status"`
	SupportEmail   string             `json:"supportEmail"`
	TwitterName    string             `json:"twitterName"`
	Website        string             `json:"website"`
}
