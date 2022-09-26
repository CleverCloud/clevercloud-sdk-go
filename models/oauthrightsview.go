package models

// OAuthRightsView:
type OAuthRightsView struct {
	Access_organisations                        bool `json:"access_organisations"`
	Access_organisations_bills                  bool `json:"access_organisations_bills"`
	Access_organisations_consumption_statistics bool `json:"access_organisations_consumption_statistics"`
	Access_organisations_credit_count           bool `json:"access_organisations_credit_count"`
	Access_personal_information                 bool `json:"access_personal_information"`
	Almighty                                    bool `json:"almighty"`
	Manage_organisations                        bool `json:"manage_organisations"`
	Manage_organisations_applications           bool `json:"manage_organisations_applications"`
	Manage_organisations_members                bool `json:"manage_organisations_members"`
	Manage_organisations_services               bool `json:"manage_organisations_services"`
	Manage_personal_information                 bool `json:"manage_personal_information"`
	Manage_ssh_keys                             bool `json:"manage_ssh_keys"`
}
