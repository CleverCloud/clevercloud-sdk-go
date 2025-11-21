# Response API Changes

## Summary

The SDK has been updated to return `client.Response[T]` instead of `(T, error)` from all service functions and builder methods. This provides more control over the response handling and gives access to HTTP metadata.

## Benefits

1. **Access to HTTP metadata**: Status code, headers (including Sozu-Id), etc.
2. **Consistent error handling**: Check `response.HasError()` instead of `err != nil`
3. **Better error context**: Errors include HTTP status codes and response bodies
4. **Simplified code**: No need to handle both result and error separately

## Migration Guide

### Before (Old API)

```go
result, err := base.Listavailableresources(ctx, client, tracer, tenantId)
if err != nil {
    // Handle error
    return err
}
// Use result
fmt.Printf("Resources: %v\n", result)
```

### After (New API)

```go
response := base.Listavailableresources(ctx, client, tracer, tenantId)
if response.HasError() {
    // Handle error
    return response.Error()
}
// Use payload
result := response.Payload()
fmt.Printf("Resources: %v\n", result)

// Access HTTP metadata
fmt.Printf("Status: %d\n", response.StatusCode())
fmt.Printf("Sozu-Id: %s\n", response.SozuID())
```

### Using the Builder Pattern

```go
sdk := sdk.NewSDK(sdk.WithClient(client), sdk.WithTracer(tracer))

// Old way (no longer works):
// resources, err := sdk.V4.Tenants.TenantId(tenantId).Resources.Listavailableresources(ctx)

// New way:
response := sdk.V4.Tenants.TenantId(tenantId).Resources.Listavailableresources(ctx)
if response.HasError() {
    log.Fatalf("Error: %v", response.Error())
}
resources := response.Payload()
```

## Response[T] Interface

The `client.Response[T]` interface provides:

```go
type Response[T any] interface {
    // Error handling
    Error() error
    HasError() bool
    IsNotFoundError() bool

    // HTTP metadata
    StatusCode() int
    SozuID() string

    // Data access
    Payload() *T

    // Comparison
    Equal(anotherResponse Response[T]) bool
}
```

## Examples

### Checking for specific errors

```go
response := sdk.V4.Tenants.TenantId(tenantId).Resources.Listavailableresources(ctx)

if response.IsNotFoundError() {
    fmt.Println("Tenant not found")
    return nil
}

if response.HasError() {
    fmt.Printf("API error (status %d): %v\n", response.StatusCode(), response.Error())
    return response.Error()
}

resources := response.Payload()
```

### Using Sozu-Id for debugging

```go
response := sdk.V4.Pulsar.Addons.AddonId(addonId).PersistentTopics.List(ctx)

if response.HasError() {
    log.Printf("Request failed (Sozu-Id: %s): %v", response.SozuID(), response.Error())
    return response.Error()
}

log.Printf("Request successful (Sozu-Id: %s)", response.SozuID())
topics := response.Payload()
```

### Comparing responses

```go
response1 := sdk.V4.Pulsar.Addons.AddonId(addonId).Get(ctx)
time.Sleep(1 * time.Second)
response2 := sdk.V4.Pulsar.Addons.AddonId(addonId).Get(ctx)

if response1.Equal(response2) {
    fmt.Println("Responses are identical")
}
```
