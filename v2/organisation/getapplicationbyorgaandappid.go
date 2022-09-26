package organisation

import (
	"context"
	client "go.clever-cloud.dev/client"
	sdk "go.clever-cloud.dev/sdk"
	models "go.clever-cloud.dev/sdk/models"
)

// GetApplication:
// Tags: Organisation
func GetApplication(ctx context.Context, cc *client.Client, id string, appId string) client.Response[models.ApplicationView] {
	url := sdk.Tpl("/v2/organisations/{id}/applications/{appId}", map[string]string{
		"appId": appId,
		"id":    id,
	})
	res := client.Get[models.ApplicationView](ctx, cc, url)
	return res
}
