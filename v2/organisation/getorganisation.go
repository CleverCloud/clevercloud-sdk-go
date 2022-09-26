package organisation

import (
	"context"
	client "go.clever-cloud.dev/client"
	sdk "go.clever-cloud.dev/sdk"
	models "go.clever-cloud.dev/sdk/models"
)

// Get:
// Tags: Organisation
func Get(ctx context.Context, cc *client.Client, id string) client.Response[models.OrganisationView] {
	url := sdk.Tpl("/v2/organisations/{id}", map[string]string{"id": id})
	res := client.Get[models.OrganisationView](ctx, cc, url)
	return res
}
