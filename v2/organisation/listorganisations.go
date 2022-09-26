package organisation

import (
	"context"
	client "go.clever-cloud.dev/client"
	sdk "go.clever-cloud.dev/sdk"
	models "go.clever-cloud.dev/sdk/models"
)

// List:
// Tags: Organisation
func List(ctx context.Context, cc *client.Client) client.Response[[]models.OrganisationView] {
	url := sdk.Tpl("/v2/organisations", map[string]string{})
	res := client.Get[[]models.OrganisationView](ctx, cc, url)
	return res
}
