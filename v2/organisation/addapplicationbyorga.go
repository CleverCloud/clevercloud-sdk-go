package organisation

import (
	"context"
	client "go.clever-cloud.dev/client"
	sdk "go.clever-cloud.dev/sdk"
	models "go.clever-cloud.dev/sdk/models"
)

// CreateApplication:
// Tags: Organisation
func CreateApplication(ctx context.Context, cc *client.Client, id string, req models.WannabeApplication) client.Response[models.ApplicationView] {
	url := sdk.Tpl("/v2/organisations/{id}/applications", map[string]string{"id": id})
	res := client.Post[models.ApplicationView](ctx, cc, url, req)
	return res
}
