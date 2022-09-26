package models

// AddonTardis:
type AddonTardis struct {
	Creation_date       string                    `json:"creation_date"`
	Deletion_date       string                    `json:"deletion_date"`
	Owner_id            OwnerID                   `json:"owner_id"`
	Plan                AddonTardisPlanIDentifier `json:"plan"`
	Read_token          TardisTokenLine           `json:"read_token"`
	Senx_application_id SenXApplicationID         `json:"senx_application_id"`
	Status              AddonStatus               `json:"status"`
	Tardis_id           TardisID                  `json:"tardis_id"`
	Write_token         TardisTokenLine           `json:"write_token"`
}
