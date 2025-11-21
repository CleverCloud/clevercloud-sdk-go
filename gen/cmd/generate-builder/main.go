// Package main implements a code generator that creates builder pattern code from OpenAPI specification
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/format"
	"log"
	"os"
	"sort"
	"strings"

	. "github.com/dave/jennifer/jen"
	"github.com/swaggest/openapi-go/openapi31"
	"gopkg.in/yaml.v3"
)

// Config holds the generator configuration
type Config struct {
	SpecFile    string
	OutputFile  string
	PackageName string
}

// BuilderOperation represents an API operation
type BuilderOperation struct {
	Name         string
	Path         string
	Method       string
	Package      string
	PathParams   []Param
	QueryParams  bool
	RequestBody  bool
	RequestType  string
	ResponseType string
}

type Param struct {
	Name string
	Type string
}

// BuilderNode represents a node in the builder hierarchy
type BuilderNode struct {
	Segment    string
	IsParam    bool
	ParamName  string
	ParamType  string
	Children   map[string]*BuilderNode
	Operations []BuilderOperation
}

func main() {
	config := parseFlags()

	log.Printf("🚀 Starting builder generation from %s", config.SpecFile)

	// Parse OpenAPI specification
	spec, err := parseOpenAPISpec(config.SpecFile)
	if err != nil {
		log.Fatalf("Failed to parse OpenAPI spec: %v", err)
	}

	// Extract operations from spec
	operations := extractOperations(spec)
	log.Printf("Found %d operations", len(operations))

	// Build path tree
	root := buildPathTree(operations)

	// Generate builder code
	if err := generateBuilderCode(root, config); err != nil {
		log.Fatalf("Failed to generate builder code: %v", err)
	}

	log.Printf("✅ Successfully generated builder code in %s", config.OutputFile)
}

func parseFlags() *Config {
	config := &Config{}

	flag.StringVar(&config.SpecFile, "spec", "openapi.json", "Path to OpenAPI specification file")
	flag.StringVar(&config.OutputFile, "output", "../../builder.go", "Output file for generated builder")
	flag.StringVar(&config.PackageName, "package", "sdk", "Go package name")
	flag.Parse()

	return config
}

func parseOpenAPISpec(specFile string) (*openapi31.Spec, error) {
	data, err := os.ReadFile(specFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read spec file: %w", err)
	}

	spec := &openapi31.Spec{}

	// Try JSON first
	if err := json.Unmarshal(data, spec); err != nil {
		// Try YAML
		if err := yaml.Unmarshal(data, spec); err != nil {
			return nil, fmt.Errorf("failed to parse spec as JSON or YAML: %w", err)
		}
	}

	return spec, nil
}

func extractOperations(spec *openapi31.Spec) []BuilderOperation {
	var operations []BuilderOperation

	if spec.Paths == nil {
		return operations
	}

	for path, pathItem := range spec.Paths.MapOfPathItemValues {
		// Process each HTTP method
		methods := map[string]*openapi31.Operation{
			"GET":    pathItem.Get,
			"POST":   pathItem.Post,
			"PUT":    pathItem.Put,
			"DELETE": pathItem.Delete,
			"PATCH":  pathItem.Patch,
		}

		for method, op := range methods {
			if op == nil {
				continue
			}

			// Apply operationId overrides for duplicate IDs
			operationID := ""
			if op.ID != nil {
				operationID = *op.ID
			}
			overrideKey := method + ":" + path
			if override, exists := OPERATION_ID_OVERRIDES[overrideKey]; exists {
				operationID = override
			}

			// Get x-service from MapOfAnything
			xService := ""
			if xs, ok := op.MapOfAnything["x-service"].(string); ok {
				xService = xs
			}

			// If no x-service, try to derive from tags or path
			if xService == "" {
				if len(op.Tags) > 0 {
					xService = op.Tags[0]
				} else {
					continue
				}
			}

			// Map to package name (check operation exceptions first, then service mapping)
			packageName := getPackageForOperation(operationID, xService, path)
			if packageName == "" {
				continue // Skip operations without a service
			}

			// Extract parameters
			var pathParams []Param
			hasQueryParams := false

			for _, paramOrRef := range op.Parameters {
				param := paramOrRef.Parameter
				if param == nil {
					continue
				}
				if string(param.In) == "path" {
					pathParams = append(pathParams, Param{
						Name: param.Name,
						Type: schemaMapToGoType(param.Schema),
					})
				} else if string(param.In) == "query" {
					hasQueryParams = true
				}
			}

			// Extract request body type
			hasRequestBody := false
			requestType := ""
			if op.RequestBody != nil && op.RequestBody.RequestBody != nil {
				for _, mediaType := range op.RequestBody.RequestBody.Content {
					// Only set hasRequestBody if we have a proper model type (with $ref)
					schemaRef := getSchemaRef(mediaType.Schema)
					if schemaRef != "" {
						requestType = schemaMapToGoType(mediaType.Schema)
						hasRequestBody = true
					} else if getSchemaType(mediaType.Schema) == "array" && getSchemaItemsRef(mediaType.Schema) != "" {
						requestType = schemaMapToGoType(mediaType.Schema)
						hasRequestBody = true
					}
					break
				}
			}

			// Extract response type - check 200, 201, 202, 204 status codes
			responseType := ""
			foundSuccessResponse := false
			if op.Responses != nil {
				for _, statusCode := range []string{"200", "201", "202", "204"} {
					if respOrRef, ok := op.Responses.MapOfResponseOrReferenceValues[statusCode]; ok {
						foundSuccessResponse = true
						resp := respOrRef.Response
						if resp == nil {
							responseType = "client.Nothing"
							break
						}
						// Check if response has content
						if len(resp.Content) == 0 {
							responseType = "client.Nothing"
							break
						}
						for _, mediaType := range resp.Content {
							responseType = schemaMapToGoTypeForResponse(mediaType.Schema)
							break
						}
						if responseType != "" && responseType != "any" {
							break
						}
					}
				}
			}
			// If we found a success response but no type, it means empty content
			if foundSuccessResponse && (responseType == "" || responseType == "any") {
				responseType = "client.Nothing"
			}

			operations = append(operations, BuilderOperation{
				Name:         operationID,
				Path:         path,
				Method:       method,
				Package:      packageName,
				PathParams:   pathParams,
				QueryParams:  hasQueryParams,
				RequestBody:  hasRequestBody,
				RequestType:  requestType,
				ResponseType: responseType,
			})
		}
	}

	return operations
}

// SERVICE_NAME_EXCEPTIONS maps x-service values to package names
// Must match generate-services/main.go SERVICE_NAME_EXCEPTIONS
var SERVICE_NAME_EXCEPTIONS = map[string]string{
	"addon-pulsar":     "pulsar",
	"addon-storage":    "storage",
	"addon-cellar":     "storage",
	"config-provider":  "configuration_provider",
	"network-group":    "network_group",
	"infrastructure":   "infrastructure",
	"compute":          "infrastructure",
	"drain":            "drains",
	"drains":           "drains",
	"materia-kv":       "materia_kv",
	"addon-ai":         "ai",
	"addon-cumulocity": "cumulocity",
	"addon-keycloak":   "keycloak",
	"addon-matomo":     "matomo",
	"addon-metabase":   "metabase",
	"addon-otoroshi":   "otoroshi",
	"addon-postgresql": "postgresql",
	"loadbalancer":     "loadbalancer",
}

// OPERATION_ID_OVERRIDES fixes duplicate operationId issues in the OpenAPI spec
// Must match generate-services/main.go OPERATION_ID_OVERRIDES
var OPERATION_ID_OVERRIDES = map[string]string{
	// Fix duplicate "createEndpoint" operationId
	"DELETE:/v4/ai/organisations/{ownerId}/ai/{addonAIId}/endpoints/{endpointId}": "deleteAIEndpoint",
	"POST:/v4/ai/organisations/{ownerId}/ai/{addonAIId}/endpoints":                "createAIEndpoint",
	// Fix duplicate getMateriaKv/deleteMateriaKv (exist on both /v2 and /v4 paths)
	"GET:/v4/materia/organisations/{ownerId}/materia/databases/{resourceId}":    "getMateriaKvV4",
	"DELETE:/v4/materia/organisations/{ownerId}/materia/databases/{resourceId}": "deleteMateriaKvV4",
}

// OPERATION_MAPPING_EXCEPTIONS maps specific operations to target services
// Must match generate-services/main.go OPERATION_MAPPING_EXCEPTIONS
var OPERATION_MAPPING_EXCEPTIONS = map[string]string{
	// Pulsar operations mal classées (x-service: product → pulsar)
	"getPulsar":                      "pulsar",
	"deletePulsar":                   "pulsar",
	"getPulsarPolicies":              "pulsar",
	"setStoragePolicies":             "pulsar",
	"renewPulsarToken":               "pulsar",
	"createPulsarTenantAndNamespace": "pulsar",
	"deletePulsarTenantAndNamespace": "pulsar",
	"createPulsar":                   "pulsar",
	"getPulsarProviderInfo":          "pulsar",
	"getPulsarCluster":               "pulsar",
	"listPulsarConsumptions":         "pulsar",
	"getTriggerPulsar":               "pulsar",
	"createTriggerPulsar":            "pulsar",
	"deleteTriggerPulsar":            "pulsar",

	// V2 Pulsar operations (x-service: product → pulsar)
	"provisionPulsar": "pulsar",
	"getPulsarV2":     "pulsar",
}

func mapServiceToPackage(xService string) string {
	if pkg, ok := SERVICE_NAME_EXCEPTIONS[xService]; ok {
		return pkg
	}

	// Default: replace hyphens with underscores
	return strings.ReplaceAll(xService, "-", "_")
}

// getPackageForOperation returns the package for a specific operation, checking exceptions first
func getPackageForOperation(operationID, xService, path string) string {
	// Check operation-specific exceptions first
	if pkg, ok := OPERATION_MAPPING_EXCEPTIONS[operationID]; ok {
		return pkg
	}

	// Apply mapping rules (must match generate-services applyMappingRules)
	if pkg := applyMappingRules(operationID, path); pkg != "" {
		return pkg
	}

	// Fall back to service-level mapping
	return mapServiceToPackage(xService)
}

// applyMappingRules applies path-based and operationId-based mapping rules
// Must match generate-services/main.go applyMappingRules
func applyMappingRules(operationID, path string) string {
	// Rule 1: Toute opération avec "pulsar" dans le path → service pulsar
	if strings.Contains(strings.ToLower(path), "pulsar") {
		return "pulsar"
	}

	// Rule 2: Toute opération avec "Pulsar" dans l'operationId → service pulsar
	if strings.Contains(operationID, "Pulsar") || strings.Contains(strings.ToLower(operationID), "pulsar") {
		return "pulsar"
	}

	// No mapping rule applies
	return ""
}

// Helper functions to extract values from map[string]any schemas
func getSchemaRef(schema map[string]any) string {
	if ref, ok := schema["$ref"].(string); ok {
		return ref
	}
	return ""
}

func getSchemaType(schema map[string]any) string {
	if t, ok := schema["type"].(string); ok {
		return t
	}
	return ""
}

func getSchemaFormat(schema map[string]any) string {
	if f, ok := schema["format"].(string); ok {
		return f
	}
	return ""
}

func getSchemaItemsRef(schema map[string]any) string {
	if items, ok := schema["items"].(map[string]any); ok {
		return getSchemaRef(items)
	}
	return ""
}

func schemaMapToGoType(schema map[string]any) string {
	// Handle $ref
	if ref := getSchemaRef(schema); ref != "" {
		parts := strings.Split(ref, "/")
		if len(parts) > 0 {
			modelName := parts[len(parts)-1]
			switch strings.ToLower(modelName) {
			case "string", "int", "int64", "float64", "bool", "any":
				return strings.ToLower(modelName)
			default:
				return "*models." + toGoStructName(modelName)
			}
		}
	}

	// Handle array
	schemaType := getSchemaType(schema)
	if schemaType == "array" {
		if items, ok := schema["items"].(map[string]any); ok {
			itemType := schemaMapToGoType(items)
			return "[]" + itemType
		}
	}

	// Map basic types
	switch schemaType {
	case "string":
		return "string"
	case "integer":
		if getSchemaFormat(schema) == "int64" {
			return "int64"
		}
		return "int"
	case "number":
		return "float64"
	case "boolean":
		return "bool"
	default:
		return "any"
	}
}

// schemaMapToGoTypeForResponse is like schemaMapToGoType but returns "any" for
// non-model types to match what services generate
func schemaMapToGoTypeForResponse(schema map[string]any) string {
	// Handle $ref - only single model references get typed
	if ref := getSchemaRef(schema); ref != "" {
		parts := strings.Split(ref, "/")
		if len(parts) > 0 {
			modelName := parts[len(parts)-1]
			switch strings.ToLower(modelName) {
			case "string", "int", "int64", "float64", "bool", "any":
				return "any"
			default:
				return "models." + toGoStructName(modelName)
			}
		}
	}

	// Handle array responses: { "type": "array", "items": { "$ref": "..." } }
	if getSchemaType(schema) == "array" {
		if itemsRef := getSchemaItemsRef(schema); itemsRef != "" {
			parts := strings.Split(itemsRef, "/")
			if len(parts) > 0 {
				modelName := parts[len(parts)-1]
				return "[]models." + toGoStructName(modelName)
			}
		}
	}

	// Primitives and inline objects become 'any'
	return "any"
}

func buildPathTree(operations []BuilderOperation) *BuilderNode {
	root := &BuilderNode{
		Segment:  "root",
		Children: make(map[string]*BuilderNode),
	}

	for _, op := range operations {
		// Parse path segments
		segments := strings.Split(strings.Trim(op.Path, "/"), "/")

		current := root
		for i, segment := range segments {
			// Check if this is a path parameter
			isParam := strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}")

			var key string
			var paramName string
			var paramType string

			if isParam {
				// Extract parameter name from {paramName}
				paramName = strings.Trim(segment, "{}")

				// Find parameter type from operation
				paramType = "string" // default
				for _, p := range op.PathParams {
					if p.Name == paramName {
						paramType = p.Type
						break
					}
				}

				// Use ":param" as key for all parameters at this position
				key = ":" + paramName
			} else {
				key = segment
			}

			// Get or create child node
			child, exists := current.Children[key]
			if !exists {
				child = &BuilderNode{
					Segment:   segment,
					IsParam:   isParam,
					ParamName: paramName,
					ParamType: paramType,
					Children:  make(map[string]*BuilderNode),
				}
				current.Children[key] = child
			}

			current = child

			// If this is the last segment, add the operation
			if i == len(segments)-1 {
				current.Operations = append(current.Operations, op)
			}
		}
	}

	return root
}

func generateBuilderCode(root *BuilderNode, config *Config) error {
	// Start from /v4 if it exists
	var startNode *BuilderNode
	if v4, ok := root.Children["v4"]; ok {
		startNode = v4
	} else {
		startNode = root
	}

	// Collect unique service packages
	packages := make(map[string]bool)
	collectPackages(startNode, packages)

	// Create file
	f := NewFile(config.PackageName)

	// Add file comment
	f.Comment("Code generated by generate-builder. DO NOT EDIT.")

	// Generate builders recursively
	generateBuilderNodes(f, startNode, "V4", []FieldInfo{{Name: "sdk", Type: Op("*").Id("sdkImpl")}}, packages)

	// Render to buffer first to fix Jen spacing issues
	var buf strings.Builder
	f.NoFormat = true // Get raw output without formatting
	err := f.Render(&buf)
	if err != nil {
		return fmt.Errorf("failed to render: %w", err)
	}

	// Fix spacing issues in array pointer types: "[] *" -> "[]*"
	fixed := strings.ReplaceAll(buf.String(), "] *", "]*")
	// Fix spacing issues in regular pointers: "( *" -> "(*"
	fixed = strings.ReplaceAll(fixed, "( *", "(*")
	// Fix spacing issues after commas: ",(" -> ", ("
	fixed = strings.ReplaceAll(fixed, ",(", ", (")

	// Format with gofmt
	formatted, err := format.Source([]byte(fixed))
	if err != nil {
		// Save the unformatted version for debugging
		os.WriteFile(config.OutputFile+".unformatted", []byte(fixed), 0644)
		return fmt.Errorf("failed to format (saved to %s.unformatted): %w", config.OutputFile, err)
	}

	// Write to file
	err = os.WriteFile(config.OutputFile, formatted, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func collectPackages(node *BuilderNode, packages map[string]bool) {
	for _, op := range node.Operations {
		packages[op.Package] = true
	}
	for _, child := range node.Children {
		collectPackages(child, packages)
	}
}

type FieldInfo struct {
	Name string
	Type *Statement
}

func generateBuilderNodes(f *File, node *BuilderNode, prefix string, parentFields []FieldInfo, packages map[string]bool) {
	builderName := prefix + "Builder"
	implName := strings.ToLower(string(prefix[0])) + prefix[1:] + "BuilderImpl"

	// Start with parent fields
	fieldInfos := make([]FieldInfo, len(parentFields))
	copy(fieldInfos, parentFields)

	// Collect methods for interface
	var interfaceMethods []Code

	// Methods for child nodes
	for _, child := range sortedChildren(node.Children) {
		childPrefix := prefix + toPascalCase(child.Segment)

		if child.IsParam {
			// Parameter method
			paramName := toCamelCase(child.ParamName)
			paramType := parseGoType(child.ParamType)

			interfaceMethods = append(interfaceMethods,
				Id(toPascalCase(child.ParamName)).Params(Id(paramName).Add(paramType)).Id(childPrefix+"Builder"))
		} else {
			// Static segment method
			interfaceMethods = append(interfaceMethods,
				Id(toPascalCase(child.Segment)).Params().Id(childPrefix+"Builder"))
		}
	}

	// Methods for operations
	for _, op := range node.Operations {
		// Build method signature
		params := []Code{Id("ctx").Qual("context", "Context")}

		if op.RequestBody {
			reqType := parseModelType(op.RequestType)
			params = append(params, Id("request").Add(reqType))
		}

		if op.QueryParams {
			params = append(params, Id("opts").Op("...").Qual("go.clever-cloud.dev/sdk/services/"+op.Package, "Option"))
		}

		// Build return type as client.Response[T]
		var returnType *Statement
		if op.ResponseType != "" && op.ResponseType != "unknown" && op.ResponseType != "any" {
			retType := parseModelTypeForResponse(op.ResponseType)
			returnType = Qual("go.clever-cloud.dev/client", "Response").Types(retType)
		} else {
			returnType = Qual("go.clever-cloud.dev/client", "Response").Types(Any())
		}

		// Add method to interface
		method := Id(toPascalCase(op.Name)).Params(params...).Add(returnType)
		interfaceMethods = append(interfaceMethods, method)
	}

	// Generate interface
	f.Comment(fmt.Sprintf("%s provides access to operations", builderName))
	f.Type().Id(builderName).Interface(interfaceMethods...)
	f.Line()

	// Generate implementation struct
	structFields := []Code{}
	for _, field := range fieldInfos {
		structFields = append(structFields, Id(field.Name).Add(field.Type))
	}
	f.Comment(fmt.Sprintf("%s implements %s", implName, builderName))
	f.Type().Id(implName).Struct(structFields...)
	f.Line()

	// Generate constructor
	constructorParams := []Code{}
	for _, field := range fieldInfos {
		constructorParams = append(constructorParams, Id(field.Name).Add(field.Type))
	}

	constructorBody := Dict{}
	for _, field := range fieldInfos {
		constructorBody[Id(field.Name)] = Id(field.Name)
	}

	f.Comment(fmt.Sprintf("new%s creates a new %s", builderName, builderName))
	f.Func().Id("new" + builderName).Params(constructorParams...).Id(builderName).Block(
		Return(Op("&").Id(implName).Values(constructorBody)))
	f.Line()

	// Generate method implementations for child nodes
	for _, child := range sortedChildren(node.Children) {
		childPrefix := prefix + toPascalCase(child.Segment)

		if child.IsParam {
			// Parameter method
			paramName := toCamelCase(child.ParamName)
			paramType := parseGoType(child.ParamType)

			// Build call args - skip field if it has the same name as the new param
			// (handles duplicate param names in paths like /databases/{objectId}/schemas/{objectId})
			callArgs := []Code{}
			for _, field := range fieldInfos {
				if field.Name != paramName {
					callArgs = append(callArgs, Id("b").Dot(field.Name))
				}
			}
			callArgs = append(callArgs, Id(paramName))

			f.Comment(fmt.Sprintf("%s returns builder for %s", toPascalCase(child.ParamName), paramName))
			f.Func().Params(Id("b").Op("*").Id(implName)).Id(toPascalCase(child.ParamName)).
				Params(Id(paramName).Add(paramType)).Id(childPrefix + "Builder").Block(
				Return(Id("new" + childPrefix + "Builder").Call(callArgs...)))
			f.Line()
		} else {
			// Static segment method
			callArgs := []Code{}
			for _, field := range fieldInfos {
				callArgs = append(callArgs, Id("b").Dot(field.Name))
			}

			f.Comment(fmt.Sprintf("%s returns %s builder", toPascalCase(child.Segment), toPascalCase(child.Segment)))
			f.Func().Params(Id("b").Op("*").Id(implName)).Id(toPascalCase(child.Segment)).
				Params().Id(childPrefix + "Builder").Block(
				Return(Id("new" + childPrefix + "Builder").Call(callArgs...)))
			f.Line()
		}
	}

	// Generate method implementations for operations
	for _, op := range node.Operations {
		// Build method signature
		params := []Code{Id("ctx").Qual("context", "Context")}
		callParams := []Code{Id("ctx"), Id("b").Dot("sdk").Dot("Client").Call(), Id("b").Dot("sdk").Dot("Tracer").Call()}

		// Add path params from stored fields (skip sdk field)
		for _, field := range fieldInfos {
			if field.Name != "sdk" {
				callParams = append(callParams, Id("b").Dot(field.Name))
			}
		}

		if op.RequestBody {
			reqType := parseModelType(op.RequestType)
			params = append(params, Id("request").Add(reqType))
			callParams = append(callParams, Id("request"))
		}

		if op.QueryParams {
			params = append(params, Id("opts").Op("...").Qual("go.clever-cloud.dev/sdk/services/"+op.Package, "Option"))
			callParams = append(callParams, Id("opts").Op("..."))
		}

		// Build return type as client.Response[T]
		var returnType *Statement
		if op.ResponseType != "" && op.ResponseType != "unknown" && op.ResponseType != "any" {
			retType := parseModelTypeForResponse(op.ResponseType)
			returnType = Qual("go.clever-cloud.dev/client", "Response").Types(retType)
		} else {
			returnType = Qual("go.clever-cloud.dev/client", "Response").Types(Any())
		}

		f.Comment(fmt.Sprintf("%s calls %s.%s", toPascalCase(op.Name), op.Package, toPascalCase(op.Name)))
		f.Func().Params(Id("b").Op("*").Id(implName)).Id(toPascalCase(op.Name)).
			Params(params...).Add(returnType).Block(
			Return(Qual("go.clever-cloud.dev/sdk/services/"+op.Package, toPascalCase(op.Name)).Call(callParams...)))
		f.Line()
	}

	// Recursively generate child builders
	for _, child := range sortedChildren(node.Children) {
		childPrefix := prefix + toPascalCase(child.Segment)

		// Prepare fields for child builder
		childFields := make([]FieldInfo, len(fieldInfos))
		copy(childFields, fieldInfos)

		// If child is a parameter, add it to the fields
		if child.IsParam {
			paramName := toCamelCase(child.ParamName)
			fieldExists := false
			for _, existing := range childFields {
				if existing.Name == paramName {
					fieldExists = true
					break
				}
			}
			if !fieldExists {
				childFields = append(childFields, FieldInfo{
					Name: paramName,
					Type: parseGoType(child.ParamType),
				})
			}
		}

		generateBuilderNodes(f, child, childPrefix, childFields, packages)
	}
}

func sortedChildren(children map[string]*BuilderNode) []*BuilderNode {
	var keys []string
	for k := range children {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var result []*BuilderNode
	for _, k := range keys {
		result = append(result, children[k])
	}
	return result
}

func parseGoType(typeName string) *Statement {
	// Handle pointer types
	if after, ok := strings.CutPrefix(typeName, "*"); ok {
		return Op("*").Add(parseGoType(after))
	}

	// Handle array types
	if elementType, ok := strings.CutPrefix(typeName, "[]"); ok {
		return Index().Add(parseGoType(elementType))
	}

	// Map basic types
	switch typeName {
	case "string":
		return String()
	case "int":
		return Int()
	case "int64":
		return Int64()
	case "float64":
		return Float64()
	case "bool":
		return Bool()
	case "any":
		return Any()
	default:
		return Id(typeName)
	}
}

func parseModelType(typeName string) *Statement {
	if typeName == "" || typeName == "unknown" || typeName == "any" {
		return Any()
	}

	// Handle array types - must handle before removing pointer
	if elementType, ok := strings.CutPrefix(typeName, "[]"); ok {

		// Handle []*models.Something
		if modelName, ok0 := strings.CutPrefix(elementType, "*models."); ok0 {
			return Index().Op("*").Qual("go.clever-cloud.dev/sdk/models", modelName)
		}

		// Handle []models.Something
		if modelName, ok0 := strings.CutPrefix(elementType, "models."); ok0 {
			return Index().Op("*").Qual("go.clever-cloud.dev/sdk/models", modelName)
		}

		// Handle []*Something (assuming it's a model)
		if modelName, ok0 := strings.CutPrefix(elementType, "*"); ok0 {
			// Check if it's a basic type
			switch modelName {
			case "string", "int", "int64", "float64", "bool", "any":
				return Index().Add(parseGoType(modelName))
			default:
				return Index().Op("*").Qual("go.clever-cloud.dev/sdk/models", modelName)
			}
		}

		// Handle []string or other basic types
		return Index().Add(parseGoType(elementType))
	}

	// Remove pointer prefix if present
	isPointer := strings.HasPrefix(typeName, "*")
	if isPointer {
		typeName = strings.TrimPrefix(typeName, "*")
	}

	// Remove models. prefix if present
	typeName = strings.TrimPrefix(typeName, "models.")

	if typeName == "any" {
		return Any()
	}

	// Check if it's a basic type
	switch typeName {
	case "string", "int", "int64", "float64", "bool":
		if isPointer {
			return Op("*").Add(parseGoType(typeName))
		}
		return parseGoType(typeName)
	}

	// For models types, always return pointer
	// Apply same Go struct name conversion as model generator
	return Op("*").Qual("go.clever-cloud.dev/sdk/models", toGoStructName(typeName))
}

// parseModelTypeForResponse is like parseModelType but doesn't add pointers
// Used for Response[T] where T should match the service return type (without pointer)
func parseModelTypeForResponse(typeName string) *Statement {
	if typeName == "" || typeName == "unknown" || typeName == "any" {
		return Any()
	}

	// Handle client.Nothing for empty responses (204, etc.)
	if typeName == "client.Nothing" {
		return Qual("go.clever-cloud.dev/client", "Nothing")
	}

	// Handle array types
	if elementType, ok := strings.CutPrefix(typeName, "[]"); ok {
		if modelName, ok0 := strings.CutPrefix(elementType, "*models."); ok0 {
			return Index().Op("*").Qual("go.clever-cloud.dev/sdk/models", toGoStructName(modelName))
		}
		if modelName, ok0 := strings.CutPrefix(elementType, "models."); ok0 {
			return Index().Qual("go.clever-cloud.dev/sdk/models", toGoStructName(modelName))
		}
		if modelName, ok0 := strings.CutPrefix(elementType, "*"); ok0 {
			switch modelName {
			case "string", "int", "int64", "float64", "bool", "any":
				return Index().Add(parseGoType(modelName))
			default:
				return Index().Op("*").Qual("go.clever-cloud.dev/sdk/models", toGoStructName(modelName))
			}
		}
		return Index().Add(parseGoType(elementType))
	}

	// Remove pointer prefix if present
	isPointer := strings.HasPrefix(typeName, "*")
	if isPointer {
		typeName = strings.TrimPrefix(typeName, "*")
	}

	// Remove models. prefix if present
	typeName = strings.TrimPrefix(typeName, "models.")

	if typeName == "any" {
		return Any()
	}

	// Check if it's a basic type
	switch typeName {
	case "string", "int", "int64", "float64", "bool":
		return parseGoType(typeName)
	}

	// For models types, DON'T add pointer (services return value types in Response[T])
	return Qual("go.clever-cloud.dev/sdk/models", toGoStructName(typeName))
}

func toPascalCase(s string) string {
	// Remove special characters
	s = strings.Trim(s, "{}")
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, ".", "_")

	// Split by underscore
	parts := strings.Split(s, "_")

	var result string
	for _, part := range parts {
		if len(part) > 0 {
			result += strings.ToUpper(string(part[0])) + strings.ToLower(part[1:])
		}
	}

	return fixIDSuffixes(result)
}

// toGoStructName converts OpenAPI schema names to Go struct names
// Must match the logic in generate-models/main.go
func toGoStructName(name string) string {
	// Preserve exact casing for certain names to avoid conflicts
	if name == "WireGuard" {
		return "WireGuard"
	}
	return toPascalCaseForTypes(name)
}

// toPascalCaseForTypes converts schema names to PascalCase for type names
// Must match the logic in generate-models/main.go
func toPascalCaseForTypes(s string) string {
	// Handle special cases and clean input
	s = strings.ReplaceAll(s, ".", "_DOT_")
	s = strings.ReplaceAll(s, "+", "_PLUS_")
	s = strings.ReplaceAll(s, "*", "_STAR_")
	s = strings.ReplaceAll(s, "/", "_SLASH_")

	// Check if string contains separators (underscore, hyphen, space)
	hasSeparators := strings.ContainsAny(s, "_- ")

	if !hasSeparators {
		// No separators - assume already in PascalCase or single word
		// Just ensure first letter is uppercase
		if len(s) > 0 {
			return strings.ToUpper(s[:1]) + s[1:]
		}
		return s
	}

	// Split on underscores, hyphens, and spaces
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == ' '
	})

	for i, part := range parts {
		if len(part) > 0 {
			// Capitalize first letter of each part
			parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
		}
	}

	result := strings.Join(parts, "")

	// Ensure it starts with a letter or underscore (valid Go identifier)
	if len(result) > 0 && (result[0] >= '0' && result[0] <= '9') {
		result = "_" + result
	}

	return result
}

func toCamelCase(s string) string {
	// Remove special characters and convert to camelCase
	s = strings.Trim(s, "{}")
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, ".", "_")

	// Split by underscore
	parts := strings.Split(s, "_")

	result := ""
	for i, part := range parts {
		if part == "" {
			continue
		}
		if i == 0 {
			// First part stays lowercase
			result += strings.ToLower(part)
		} else {
			// Subsequent parts capitalize first letter
			result += strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
		}
	}

	result = fixIDSuffixes(result)

	// Check for reserved keywords and append underscore
	reserved := map[string]bool{
		"break": true, "case": true, "chan": true, "const": true, "continue": true,
		"default": true, "defer": true, "else": true, "fallthrough": true, "for": true,
		"func": true, "go": true, "goto": true, "if": true, "import": true,
		"interface": true, "map": true, "package": true, "range": true, "return": true,
		"select": true, "struct": true, "switch": true, "type": true, "var": true,
	}
	if reserved[result] {
		result += "_"
	}

	return result
}

func fixIDSuffixes(s string) string {
	replacements := map[string]string{
		"Id":    "ID",
		"Url":   "URL",
		"Api":   "API",
		"Http":  "HTTP",
		"Https": "HTTPS",
		"Json":  "JSON",
		"Xml":   "XML",
		"Yaml":  "YAML",
	}

	for old, new := range replacements {
		s = strings.ReplaceAll(s, old, new)
	}

	return s
}
