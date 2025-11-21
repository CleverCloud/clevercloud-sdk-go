# OpenAPI Parsing Strategy

## Decision: Custom Structures vs swaggest/openapi-go

### Context

We evaluated using [swaggest/openapi-go](https://github.com/swaggest/openapi-go) for parsing OpenAPI specifications in our code generators but decided to use custom, simplified structures instead.

### Analysis of swaggest/openapi-go

**Strengths:**
- Full-featured OpenAPI 3.0.x and 3.1 support
- Bidirectional: Can both parse and generate OpenAPI specs
- Type-safe with proper Go generics
- Well-maintained library
- Handles complex scenarios like `$ref` resolution

**Challenges for Our Use Case:**
- Uses complex wrapper types (`ParameterOrRef`, `SchemaOrRef`, `ResponseOrRef`) for `$ref` handling
- More verbose to work with due to pointer fields and optional types
- Designed for bidirectional operations (we only need parsing)
- Additional abstraction layers (e.g., `MapOfPathItemValues`) complicate traversal
- Learning curve for developers maintaining the generators

### Our Custom Approach

We implemented minimal structures that:

1. **Parse exactly what we need**: Only the fields required for code generation
2. **Simple field access**: Direct field access without wrapper types
3. **Readable code**: Easy to understand and maintain
4. **JSON/YAML support**: Works with both formats via standard library
5. **No external dependencies**: Beyond standard library and yaml.v3

### Example Comparison

**With swaggest/openapi-go:**
```go
// Complex access pattern
for path, pathItem := range spec.Paths.MapOfPathItemValues {
    if pathItem.Get != nil {
        op := pathItem.Get
        if op.OperationID != nil {
            // Need to handle Operation vs OperationOrRef
            // Parameters are []ParameterOrRef requiring unwrapping
            for _, paramRef := range op.Parameters {
                if param := paramRef.Parameter; param != nil {
                    // Access param.In, param.Name, etc.
                    // Schema is *SchemaOrRef requiring more unwrapping
                }
            }
        }
    }
}
```

**With our custom structures:**
```go
// Simple, direct access
for path, pathItem := range spec.Paths {
    if pathItem.Get != nil {
        op := pathItem.Get
        for _, param := range op.Parameters {
            if param.In == "path" {
                // Direct field access
                name := param.Name
                schema := param.Schema
            }
        }
    }
}
```

### Custom Structure Definition

```go
type OpenAPISpec struct {
    Paths map[string]PathItem `yaml:"paths" json:"paths"`
}

type PathItem struct {
    Get    *Operation `yaml:"get,omitempty" json:"get,omitempty"`
    Post   *Operation `yaml:"post,omitempty" json:"post,omitempty"`
    Put    *Operation `yaml:"put,omitempty" json:"put,omitempty"`
    Delete *Operation `yaml:"delete,omitempty" json:"delete,omitempty"`
    Patch  *Operation `yaml:"patch,omitempty" json:"patch,omitempty"`
}

type Operation struct {
    OperationID string              `yaml:"operationId" json:"operationId"`
    Tags        []string            `yaml:"tags,omitempty" json:"tags,omitempty"`
    Parameters  []Parameter         `yaml:"parameters,omitempty" json:"parameters,omitempty"`
    RequestBody *RequestBody        `yaml:"requestBody,omitempty" json:"requestBody,omitempty"`
    Responses   map[string]Response `yaml:"responses" json:"responses"`
    Extensions  map[string]any      `yaml:",inline" json:"-"`
}

type Parameter struct {
    Name     string `yaml:"name" json:"name"`
    In       string `yaml:"in" json:"in"`
    Required bool   `yaml:"required,omitempty" json:"required,omitempty"`
    Schema   Schema `yaml:"schema" json:"schema"`
}

type Schema struct {
    Type   string  `yaml:"type,omitempty" json:"type,omitempty"`
    Format string  `yaml:"format,omitempty" json:"format,omitempty"`
    Ref    string  `yaml:"$ref,omitempty" json:"$ref,omitempty"`
    Items  *Schema `yaml:"items,omitempty" json:"items,omitempty"`
}
```

### Trade-offs

**What we gain:**
- ✅ Simpler code (~50% less complexity)
- ✅ Easier to understand and maintain
- ✅ No learning curve for contributors
- ✅ Direct field access without unwrapping
- ✅ Faster to iterate and modify

**What we sacrifice:**
- ❌ No automatic `$ref` resolution (we don't need it - we extract refs as strings)
- ❌ No validation of OpenAPI spec structure (we assume valid input)
- ❌ Limited to fields we explicitly define (sufficient for our needs)
- ❌ Can't generate OpenAPI specs (not a requirement)

### When to Reconsider

We should reconsider using swaggest/openapi-go if:

1. We need to **generate** OpenAPI specifications (not just parse)
2. We need to **resolve `$ref` references** automatically
3. We need to **validate** OpenAPI spec compliance
4. The OpenAPI spec structure changes significantly and maintaining custom types becomes burdensome
5. We want to support OpenAPI 3.1 features that differ from 3.0

### Conclusion

For our code generation use case, custom structures are the pragmatic choice. They provide exactly what we need with minimal complexity. The code is more maintainable and easier to understand for contributors.

If requirements change, migrating to swaggest/openapi-go is straightforward since we've isolated OpenAPI parsing in dedicated functions.

---

*Last updated: 2025-11-20*
*Decision made during SDK refactoring to use `client.Response[T]` return types*
