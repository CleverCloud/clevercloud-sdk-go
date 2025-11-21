package sdk_test

import (
	"testing"

	"go.clever-cloud.dev/client"
	sdk "go.clever-cloud.dev/sdk"
)

func TestGetNetworkGroup(t *testing.T) {
	s := sdk.NewSDK(sdk.WithClient(client.New(client.WithAutoOauthConfig())))

	ctx := t.Context()
	orgID := "orga_xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx" // Replace with actual org ID

	// Use builder pattern to get a network group
	response := s.V4().Kubernetes().Organisations().Ownerid(orgID).Clusters().Getclusters(ctx)
	if response.HasError() {
		t.Fatalf("Error: %v", response.Error())
	}

	ngs := *response.Payload()
	for _, ng := range ngs {
		t.Logf("Cluster %+v", ng)
	}
}
