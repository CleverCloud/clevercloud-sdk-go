package organisation

import (
	"context"
	client "go.clever-cloud.dev/client"
	sdk "go.clever-cloud.dev/sdk"
	models "go.clever-cloud.dev/sdk/models"
)

// DeleteApplication:
// Tags: Organisation
func DeleteApplication(ctx context.Context, cc *client.Client, id string, appId string) client.Response[models.Message] {
	url := sdk.Tpl("/v2/organisations/{id}/applications/{appId}", map[string]string{
		"appId": appId,
		"id":    id,
	})
	res := client.Delete[models.Message](ctx, cc, url)
	return res
}
