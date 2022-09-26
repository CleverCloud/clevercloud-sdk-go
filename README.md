# Clever Cloud SDK for Golang

[![Go Reference](https://pkg.go.dev/badge/go.clever-cloud.dev/sdk.svg)](https://pkg.go.dev/go.clever-cloud.dev/sdk)

> Most of the code is generated, you can open issues, but do not try to edit the code.


Exemple:
```go

import (
	"context"
	
	"go.clever-cloud.dev/client"
	"go.clever-cloud.dev/sdk/v2/organisation"
)

func main() {
	ctx := context.Background()
	c := client.New(client.WithAutoOauthConfig())
		
	orgRes := organisation.List(ctx, cc)
	if orgRes.HasError() {
	    panic(orgRes.Error())
	}
	
	for _, org := range *orgRes.Payload() {
	    fmt.Println(org.ID)
	}
}

```