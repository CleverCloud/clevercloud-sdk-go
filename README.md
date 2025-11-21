# Clever Cloud Go SDK

[![Go Reference](https://pkg.go.dev/badge/go.clever-cloud.dev/sdk.svg)](https://pkg.go.dev/go.clever-cloud.dev/sdk)

A Go SDK for the [Clever Cloud](https://clever-cloud.com) platform, providing a type-safe interface to interact with the Clever Cloud API.

## Installation

```bash
go get go.clever-cloud.dev/sdk
```

## Features

- **Two API levels**: Direct function calls and fluent builder pattern
- **Type-safe**: Generated from OpenAPI specification with proper Go types
- **OpenTelemetry integration**: Built-in tracing for observability
- **Functional options**: Flexible configuration for SDK and query parameters

## Usage

### Basic Setup

```go
package main

import (
    "context"

    "go.clever-cloud.dev/client"
    sdk "go.clever-cloud.dev/sdk"
)

func main() {
    // Create SDK with auto-auth (reads from ~/.config/clever-cloud/clever-tools.json)
    s := sdk.NewSDK(sdk.WithClient(client.New(client.WithAutoOauthConfig())))

    ctx := context.Background()

    // Use the SDK...
}
```

### Level 1: Direct Function API

Operations are grouped by service in `services/{service}/` packages:

```go
import "go.clever-cloud.dev/sdk/services/kubernetes"

response := kubernetes.GetClusters(ctx, client, tracer, ownerID)
if response.HasError() {
    log.Fatal(response.Error())
}
clusters := response.Payload()
```

### Level 2: Builder Pattern API

Fluent interface following URL path structure:

```go
s := sdk.NewSDK(sdk.WithClient(client.New(client.WithAutoOauthConfig())))

response := s.V4().Kubernetes().Organisations().Ownerid(orgID).Clusters().Getclusters(ctx)
if response.HasError() {
    log.Fatal(response.Error())
}
clusters := response.Payload()
```

### Query Parameters

Use functional options for query parameters:

```go
import "go.clever-cloud.dev/sdk/services/network_group"

response := network_group.GetNetworkGroups(ctx, client, tracer, ownerID,
    network_group.WithQuery("my-network"),
)
```

## Authentication

The SDK uses the [clever-cloud-client-go](https://github.com/CleverCloud/clever-cloud-client-go) for authentication. Configure credentials via:

1. **Auto-config**: Reads from `~/.config/clever-cloud/clever-tools.json` (created by clever-tools CLI)
2. **Manual config**: Pass OAuth credentials directly

```go
// Auto-config (recommended)
client.New(client.WithAutoOauthConfig())

// Manual config
client.New(client.WithOAuth1Config(consumerKey, consumerSecret, token, tokenSecret))
```

## Project Structure

```
├── sdk.go              # Main SDK entry point
├── builder.go          # Builder pattern implementation
├── models/             # Generated data structures
└── services/           # Generated API operations by service
    ├── kubernetes/
    ├── pulsar/
    ├── postgresql/
    └── ...
```

## Code Generation

This SDK is generated from the Clever Cloud OpenAPI specification. See `gen/` directory for generators.

```bash
cd gen
make all
```

## Requirements

- Go 1.24+

## License

Apache 2.0
