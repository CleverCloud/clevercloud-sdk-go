package organisation

import (
	"context"
	client "go.clever-cloud.dev/client"
	sdk "go.clever-cloud.dev/sdk"
	models "go.clever-cloud.dev/sdk/models"
)

// ListApplications:
// Tags: Organisation
// Param: 'instanceId'
func ListApplications(ctx context.Context, cc *client.Client, id string, parameters ...sdk.Parameter) client.Response[[]models.ApplicationView] {
	url := sdk.Tpl("/v2/organisations/{id}/applications", map[string]string{"id": id})
	res := client.Get[[]models.ApplicationView](ctx, cc, url)
	return res
}
func ListApplicationsParamInstanceId(value string) sdk.Parameter {
	return sdk.Parameter{
		Name:  "instanceId",
		Value: value,
	}
}
