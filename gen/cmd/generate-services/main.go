package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	. "github.com/dave/jennifer/jen"
	"github.com/swaggest/openapi-go/openapi31"
	"gopkg.in/yaml.v3"
)

// SERVICE_NAME_EXCEPTIONS maps x-service values to specific package names
// This handles special cases where the x-service name needs to be mapped to a different package name
var SERVICE_NAME_EXCEPTIONS = map[string]string{
	// Map addon services to their dedicated packages if needed
	"addon-pulsar":  "pulsar",
	"addon-storage": "storage",
	"addon-cellar":  "storage",

	// Map config-provider to product since it's product-related
	"config-provider": "product",

	// Map network-related services appropriately
	"network-group": "networkgroup",

	// Map infrastructure-related services
	"infrastructure": "infrastructure",
	"compute":        "infrastructure",

	// Map drain services
	"drain":  "drains",
	"drains": "drains",

	// Example of consolidating related services
	// "keycloak":        "product", // Move keycloak operations to product package
	// "matomo":          "product", // Move matomo operations to product package
	// "otoroshi":        "product", // Move otoroshi operations to product package
}

// OPERATION_ID_OVERRIDES fixes duplicate operationId issues in the OpenAPI spec
// Key format: "METHOD:path" -> new operationId
var OPERATION_ID_OVERRIDES = map[string]string{
	// Fix duplicate "createEndpoint" operationId
	"DELETE:/v4/ai/organisations/{ownerId}/ai/{addonAIId}/endpoints/{endpointId}": "deleteAIEndpoint",
	"POST:/v4/ai/organisations/{ownerId}/ai/{addonAIId}/endpoints":                "createAIEndpoint",
	// Fix duplicate getMateriaKv/deleteMateriaKv (exist on both /v2 and /v4 paths)
	"GET:/v4/materia/organisations/{ownerId}/materia/databases/{resourceId}":    "getMateriaKvV4",
	"DELETE:/v4/materia/organisations/{ownerId}/materia/databases/{resourceId}": "deleteMateriaKvV4",
}

// OPERATION_MAPPING_EXCEPTIONS maps specific operations to target services
// This handles operations that should be moved to different services based on SERVICE_MAPPING_EXCEPTIONS.md
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

// Service generation structures
type ServiceOperation struct {
	OperationID       string
	FunctionName      string
	PackageName       string
	Path              string
	Method            string
	Summary           string
	Description       string
	PathParams        []ServiceParam
	QueryParams       []ServiceParam
	RequestBodyType   string
	RequestBodyGoType string // Full Go type including models. prefix and array handling
	ResponseType      string
	XService          string
	TraceSpanName     string
	HasRequestBody    bool
	HasQueryParams    bool
}

type ServiceParam struct {
	Name        string
	GoName      string
	Type        string
	Required    bool
	Description string
	IsPath      bool
	IsQuery     bool
}

func main() {
	var (
		specFile       = flag.String("spec", "openapi.json", "Path to OpenAPI specification file")
		outputDir      = flag.String("output", "services", "Output directory for generated services")
		exceptionsFile = flag.String("exceptions", "", "Path to service mapping exceptions file")
	)
	flag.Parse()

	if *specFile == "" {
		log.Fatal("Please provide -spec flag with path to OpenAPI specification")
	}

	// Load exceptions file if provided
	if *exceptionsFile != "" && fileExists(*exceptionsFile) {
		log.Printf("Loading service mapping exceptions from: %s", *exceptionsFile)
		// Note: The exceptions are currently hardcoded in OPERATION_MAPPING_EXCEPTIONS
		// In the future, we could parse the markdown file to populate the map dynamically
	}

	// Read OpenAPI spec
	data, err := os.ReadFile(*specFile)
	if err != nil {
		log.Fatalf("Failed to read spec file: %v", err)
	}

	var spec openapi31.Spec

	// Detect format based on file extension
	ext := strings.ToLower(filepath.Ext(*specFile))

	switch ext {
	case ".json":
		// Parse JSON document
		if err := json.Unmarshal(data, &spec); err != nil {
			log.Fatalf("Failed to parse OpenAPI JSON: %v", err)
		}
		log.Printf("Parsed OpenAPI JSON specification")
	case ".yaml", ".yml":
		// Parse YAML document
		if err := yaml.Unmarshal(data, &spec); err != nil {
			log.Fatalf("Failed to parse OpenAPI YAML: %v", err)
		}
		log.Printf("Parsed OpenAPI YAML specification")
	default:
		// Try to auto-detect format
		log.Printf("Unknown file extension %s, attempting auto-detection...", ext)

		// Try JSON first
		err := json.Unmarshal(data, &spec)
		if err == nil {
			log.Printf("Auto-detected JSON format")
		} else {
			// Try YAML
			if err := yaml.Unmarshal(data, &spec); err != nil {
				log.Fatalf("Failed to parse OpenAPI file (tried both JSON and YAML): %v", err)
			}
			log.Printf("Auto-detected YAML format")
		}
	}

	// Process operations
	operations := extractOperations(spec)

	if len(operations) == 0 {
		log.Println("No operations found in OpenAPI spec")
		return
	}

	// Group operations by package (x-service)
	packages := groupOperationsByPackage(operations)

	// Generate service files
	for packageName, ops := range packages {
		if err := generatePackage(*outputDir, packageName, ops); err != nil {
			log.Fatalf("Failed to generate package %s: %v", packageName, err)
		}
	}

	fmt.Printf("Generated %d service operations in %d packages\n", len(operations), len(packages))
}

func extractOperations(spec openapi31.Spec) []ServiceOperation {
	var operations []ServiceOperation

	if spec.Paths == nil {
		return operations
	}

	// Sort paths alphabetically for consistent output
	var paths []string
	for path := range spec.Paths.MapOfPathItemValues {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	for _, path := range paths {
		pathItem := spec.Paths.MapOfPathItemValues[path]
		operations = append(operations, extractOperationsFromPath(path, pathItem)...)
	}

	return operations
}

func extractOperationsFromPath(path string, pathItem openapi31.PathItem) []ServiceOperation {
	var operations []ServiceOperation

	// Process methods in sorted order for consistent output
	methodOrder := []string{"DELETE", "GET", "PATCH", "POST", "PUT"}
	methodOps := map[string]*openapi31.Operation{
		"GET":    pathItem.Get,
		"POST":   pathItem.Post,
		"PUT":    pathItem.Put,
		"DELETE": pathItem.Delete,
		"PATCH":  pathItem.Patch,
	}

	for _, method := range methodOrder {
		operation := methodOps[method]
		if operation == nil {
			continue
		}

		// Apply operationId overrides for duplicate IDs
		operationID := ""
		if operation.ID != nil {
			operationID = *operation.ID
		}
		overrideKey := method + ":" + path
		if override, exists := OPERATION_ID_OVERRIDES[overrideKey]; exists {
			operationID = override
		}

		summary := ""
		if operation.Summary != nil {
			summary = *operation.Summary
		}
		description := ""
		if operation.Description != nil {
			description = *operation.Description
		}

		op := ServiceOperation{
			OperationID:   operationID,
			FunctionName:  toPascalCase(operationID),
			Path:          convertPathToGoFormat(path),
			Method:        method,
			Summary:       summary,
			Description:   description,
			TraceSpanName: operationID,
		}

		// Check operation mapping exceptions first
		if targetService, hasException := OPERATION_MAPPING_EXCEPTIONS[operationID]; hasException {
			op.XService = targetService
			op.PackageName = getPackageNameFromService(targetService)
		} else {
			// Apply mapping rules from SERVICE_MAPPING_EXCEPTIONS.md
			targetService := applyMappingRules(operationID, path, operation.Tags)
			if targetService != "" {
				op.XService = targetService
				op.PackageName = getPackageNameFromService(targetService)
			} else {
				// Extract x-service from MapOfAnything
				if xService, exists := operation.MapOfAnything["x-service"]; exists {
					if xServiceStr, ok := xService.(string); ok {
						op.XService = xServiceStr
						op.PackageName = getPackageNameFromService(xServiceStr)
					}
				}

				// If no x-service, use first tag or derive from path
				if op.PackageName == "" {
					if len(operation.Tags) > 0 {
						tagValue := operation.Tags[0]
						op.XService = tagValue
						op.PackageName = getPackageNameFromService(tagValue)
					} else {
						op.PackageName = derivePackageFromPath(path)
						op.XService = op.PackageName
					}
				}
			}
		}

		// Extract parameters
		for _, paramOrRef := range operation.Parameters {
			param := paramOrRef.Parameter
			if param == nil {
				continue
			}
			paramName := param.Name
			// Handle Go reserved words and invalid identifiers
			if isReservedWord(paramName) {
				paramName = paramName + "Param"
			}
			// Replace hyphens with underscores for valid Go identifiers
			paramName = strings.ReplaceAll(paramName, "-", "_")

			paramDesc := ""
			if param.Description != nil {
				paramDesc = *param.Description
			}
			required := false
			if param.Required != nil {
				required = *param.Required
			}

			serviceParam := ServiceParam{
				Name:        paramName,
				GoName:      toCamelCase(paramName),
				Type:        mapSchemaMapToGoType(param.Schema),
				Required:    required,
				Description: paramDesc,
				IsPath:      string(param.In) == "path",
				IsQuery:     string(param.In) == "query",
			}

			if string(param.In) == "path" {
				op.PathParams = append(op.PathParams, serviceParam)
			} else if string(param.In) == "query" {
				op.QueryParams = append(op.QueryParams, serviceParam)
				op.HasQueryParams = true
			}
		}

		// Extract request body
		if operation.RequestBody != nil && operation.RequestBody.RequestBody != nil {
			for _, mediaType := range operation.RequestBody.RequestBody.Content {
				schemaRef := getSchemaRef(mediaType.Schema)
				if schemaRef != "" {
					typeName := extractTypeFromRef(schemaRef)
					op.RequestBodyType = typeName
					op.RequestBodyGoType = "*models." + toGoStructName(typeName)
					op.HasRequestBody = true
					break
				}
				// Handle array types with items.$ref
				schemaType := getSchemaType(mediaType.Schema)
				if schemaType == "array" {
					itemsRef := getSchemaItemsRef(mediaType.Schema)
					if itemsRef != "" {
						itemType := extractTypeFromRef(itemsRef)
						op.RequestBodyType = "[]*" + itemType
						op.RequestBodyGoType = "[]*models." + toGoStructName(itemType)
						op.HasRequestBody = true
						break
					}
				}
			}
		}

		// Extract response type - check 200, 201, 202, 204 status codes
		foundSuccessResponse := false
		if operation.Responses != nil {
			for _, statusCode := range []string{"200", "201", "202", "204"} {
				if respOrRef, exists := operation.Responses.MapOfResponseOrReferenceValues[statusCode]; exists {
					foundSuccessResponse = true
					resp := respOrRef.Response
					if resp == nil {
						op.ResponseType = "NOTHING"
						break
					}
					// Check if response has content
					if len(resp.Content) == 0 {
						// No content = use client.Nothing
						op.ResponseType = "NOTHING"
						break
					}
					for _, mediaType := range resp.Content {
						schemaRef := getSchemaRef(mediaType.Schema)
						if schemaRef != "" {
							op.ResponseType = extractTypeFromRef(schemaRef)
							break
						}
						// Handle array responses
						schemaType := getSchemaType(mediaType.Schema)
						if schemaType == "array" {
							itemsRef := getSchemaItemsRef(mediaType.Schema)
							if itemsRef != "" {
								itemType := extractTypeFromRef(itemsRef)
								op.ResponseType = "[]" + itemType
								break
							}
						}
					}
					if op.ResponseType != "" {
						break
					}
				}
			}
		}
		// If we found a success response but no type, it means empty content
		if foundSuccessResponse && op.ResponseType == "" {
			op.ResponseType = "NOTHING"
		}

		operations = append(operations, op)
	}

	return operations
}

func groupOperationsByPackage(operations []ServiceOperation) map[string][]ServiceOperation {
	packages := make(map[string][]ServiceOperation)

	for _, op := range operations {
		packages[op.PackageName] = append(packages[op.PackageName], op)
	}

	return packages
}

func generatePackage(outputDir, packageName string, operations []ServiceOperation) error {
	packageDir := filepath.Join(outputDir, packageName)
	if err := os.MkdirAll(packageDir, 0755); err != nil {
		return fmt.Errorf("failed to create package directory: %v", err)
	}

	// Generate one file per operation
	for _, op := range operations {
		if err := generateOperationFile(packageDir, op); err != nil {
			return fmt.Errorf("failed to generate operation %s: %v", op.OperationID, err)
		}
	}

	// Generate options.go file for the package
	if err := generateOptionsFile(packageDir, packageName, operations); err != nil {
		return fmt.Errorf("failed to generate options.go: %v", err)
	}

	return nil
}

func generateOperationFile(packageDir string, op ServiceOperation) error {
	f := NewFile(op.PackageName)
	f.HeaderComment("Code generated by generate-services. DO NOT EDIT.")

	// Build response type
	responseType := formatResponseTypeJen(op.ResponseType)

	// Build function comment
	commentLines := []string{op.FunctionName + " " + op.Summary}
	if op.Description != "" {
		commentLines = append(commentLines, "", op.Description)
	}
	commentLines = append(commentLines, "", "Parameters:", "  - ctx: context for the request", "  - client: the Clever Cloud client", "  - tracer: OpenTelemetry tracer for observability")
	for _, p := range op.PathParams {
		commentLines = append(commentLines, fmt.Sprintf("  - %s: %s", p.Name, p.Description))
	}
	if op.HasRequestBody {
		commentLines = append(commentLines, "  - requestBody: the request payload")
	}
	if op.HasQueryParams {
		commentLines = append(commentLines, "  - opts: optional query parameters")
	}
	commentLines = append(commentLines, "", "Returns the operation result or an error", "", "Example:", "")

	// Build example call
	exampleParams := "ctx, client, tracer"
	for _, p := range op.PathParams {
		exampleParams += ", " + p.Name
	}
	if op.HasRequestBody {
		exampleParams += ", requestBody"
	}
	if op.HasQueryParams {
		exampleParams += ", opts..."
	}
	commentLines = append(commentLines,
		fmt.Sprintf("\tresponse := %s.%s(%s)", op.PackageName, op.FunctionName, exampleParams),
		"\tif response.HasError() {",
		"\t\t// Handle error",
		"\t}",
		"\tresult := response.Payload()",
		"",
		"x-service: "+op.XService,
		"operationId: "+op.OperationID,
	)

	// Build function parameters
	params := []Code{
		Id("ctx").Qual("context", "Context"),
		Id("c").Op("*").Qual("go.clever-cloud.dev/client", "Client"),
		Id("tracer").Qual("go.opentelemetry.io/otel/trace", "Tracer"),
	}
	for _, p := range op.PathParams {
		params = append(params, Id(p.Name).Id(p.Type))
	}
	if op.HasRequestBody {
		params = append(params, Id("requestBody").Add(formatRequestBodyTypeJen(op.RequestBodyGoType)))
	}
	if op.HasQueryParams {
		params = append(params, Id("opts").Op("...").Id("Option"))
	}

	// Build trace attributes
	var traceAttrs []Code
	for _, p := range op.PathParams {
		switch p.Type {
		case "string":
			traceAttrs = append(traceAttrs, Qual("go.opentelemetry.io/otel/attribute", "String").Call(Lit(p.Name), Id(p.Name)))
		case "int":
			traceAttrs = append(traceAttrs, Qual("go.opentelemetry.io/otel/attribute", "Int").Call(Lit(p.Name), Id(p.Name)))
		case "int64":
			traceAttrs = append(traceAttrs, Qual("go.opentelemetry.io/otel/attribute", "Int64").Call(Lit(p.Name), Id(p.Name)))
		case "bool":
			traceAttrs = append(traceAttrs, Qual("go.opentelemetry.io/otel/attribute", "Bool").Call(Lit(p.Name), Id(p.Name)))
		default:
			traceAttrs = append(traceAttrs, Qual("go.opentelemetry.io/otel/attribute", "String").Call(Lit(p.Name), Qual("fmt", "Sprintf").Call(Lit("%v"), Id(p.Name))))
		}
	}

	// Build path args
	pathArgs := []Code{Lit(op.Path)}
	for _, p := range op.PathParams {
		pathArgs = append(pathArgs, Id(p.Name))
	}

	// Build function body
	var body []Code

	// tracer.Start call
	tracerStartArgs := []Code{Id("ctx"), Lit(op.TraceSpanName)}
	if len(traceAttrs) > 0 {
		tracerStartArgs = append(tracerStartArgs, Qual("go.opentelemetry.io/otel/trace", "WithAttributes").Call(traceAttrs...))
	}
	body = append(body, List(Id("ctx"), Id("span")).Op(":=").Id("tracer").Dot("Start").Call(tracerStartArgs...))
	body = append(body, Defer().Id("span").Dot("End").Call())
	body = append(body, Empty())

	// path := utils.Path(...)
	body = append(body, Id("path").Op(":=").Qual("go.clever-cloud.dev/sdk/internal/utils", "Path").Call(pathArgs...))

	// Query params handling
	if op.HasQueryParams {
		body = append(body, Empty())
		body = append(body, Comment("Build query parameters"))
		body = append(body, Id("query").Op(":=").Id("buildQueryString").Call(Id("opts").Op("...")))
		body = append(body, If(Id("query").Op("!=").Lit("")).Block(
			Id("path").Op("=").Qual("fmt", "Sprintf").Call(Lit("%s?%s"), Id("path"), Id("query")),
		))
	}

	body = append(body, Empty())
	body = append(body, Comment("Make API call"))

	// API call
	var apiCall *Statement
	switch op.Method {
	case "GET":
		apiCall = Qual("go.clever-cloud.dev/client", "Get").Types(responseType).Call(Id("ctx"), Id("c"), Id("path"))
	case "POST":
		if op.HasRequestBody {
			apiCall = Qual("go.clever-cloud.dev/client", "Post").Types(responseType).Call(Id("ctx"), Id("c"), Id("path"), Id("requestBody"))
		} else {
			apiCall = Qual("go.clever-cloud.dev/client", "Post").Types(responseType).Call(Id("ctx"), Id("c"), Id("path"), Nil())
		}
	case "PUT":
		if op.HasRequestBody {
			apiCall = Qual("go.clever-cloud.dev/client", "Put").Types(responseType).Call(Id("ctx"), Id("c"), Id("path"), Id("requestBody"))
		} else {
			apiCall = Qual("go.clever-cloud.dev/client", "Put").Types(responseType).Call(Id("ctx"), Id("c"), Id("path"), Nil())
		}
	case "DELETE":
		apiCall = Qual("go.clever-cloud.dev/client", "Delete").Types(responseType).Call(Id("ctx"), Id("c"), Id("path"))
	case "PATCH":
		if op.HasRequestBody {
			apiCall = Qual("go.clever-cloud.dev/client", "Patch").Types(responseType).Call(Id("ctx"), Id("c"), Id("path"), Id("requestBody"))
		} else {
			apiCall = Qual("go.clever-cloud.dev/client", "Patch").Types(responseType).Call(Id("ctx"), Id("c"), Id("path"), Nil())
		}
	}
	body = append(body, Id("response").Op(":=").Add(apiCall))

	body = append(body, Empty())
	body = append(body, If(Id("response").Dot("HasError").Call()).Block(
		Id("span").Dot("RecordError").Call(Id("response").Dot("Error").Call()),
	))

	body = append(body, Empty())
	body = append(body, Return(Id("response")))

	// Add the function
	f.Comment(strings.Join(commentLines, "\n"))
	f.Func().Id(op.FunctionName).Params(params...).Qual("go.clever-cloud.dev/client", "Response").Types(responseType).Block(body...)

	// Write file
	filename := toSnakeCase(op.OperationID) + ".go"
	outputPath := filepath.Join(packageDir, filename)
	return f.Save(outputPath)
}

// formatResponseTypeJen returns a jennifer Code for the response type
func formatResponseTypeJen(t string) Code {
	if t == "" {
		return Id("any")
	}
	if t == "NOTHING" {
		return Qual("go.clever-cloud.dev/client", "Nothing")
	}
	if after, ok := strings.CutPrefix(t, "[]"); ok {
		return Index().Qual("go.clever-cloud.dev/sdk/models", after)
	}
	return Qual("go.clever-cloud.dev/sdk/models", t)
}

// formatRequestBodyTypeJen returns a jennifer Code for request body type like "*models.SomeType" or "[]*models.SomeType"
func formatRequestBodyTypeJen(t string) Code {
	// Handle []*models.SomeType
	if after, ok := strings.CutPrefix(t, "[]*models."); ok {
		return Index().Op("*").Qual("go.clever-cloud.dev/sdk/models", after)
	}
	// Handle []models.SomeType
	if after, ok := strings.CutPrefix(t, "[]models."); ok {
		return Index().Qual("go.clever-cloud.dev/sdk/models", after)
	}
	// Handle *models.SomeType
	if after, ok := strings.CutPrefix(t, "*models."); ok {
		return Op("*").Qual("go.clever-cloud.dev/sdk/models", after)
	}
	// Handle models.SomeType
	if after, ok := strings.CutPrefix(t, "models."); ok {
		return Qual("go.clever-cloud.dev/sdk/models", after)
	}
	// Fallback
	return Id(t)
}

func generateOptionsFile(packageDir, packageName string, operations []ServiceOperation) error {
	// Collect unique query params from all operations in this package
	queryParamMap := make(map[string]ServiceParam)
	for _, op := range operations {
		for _, qp := range op.QueryParams {
			// Use the original name as key to avoid duplicates
			if _, exists := queryParamMap[qp.Name]; !exists {
				// For options, use PascalCase for GoName and simplify complex types to string
				param := qp
				param.GoName = toPascalCase(qp.Name)
				// Query params should be simple types - convert complex refs to string
				if strings.HasPrefix(param.Type, "*") || strings.HasPrefix(param.Type, "models.") {
					param.Type = "string"
				}
				queryParamMap[qp.Name] = param
			}
		}
	}

	// Convert to sorted slice for deterministic output
	var queryParams []ServiceParam
	for _, qp := range queryParamMap {
		queryParams = append(queryParams, qp)
	}
	// Sort by name for consistent output
	sort.Slice(queryParams, func(i, j int) bool {
		return queryParams[i].Name < queryParams[j].Name
	})

	// Check if any param is a string (needs url.QueryEscape)
	hasStringParam := false
	for _, qp := range queryParams {
		if qp.Type == "string" {
			hasStringParam = true
			break
		}
	}
	_ = hasStringParam // Used in template logic below

	f := NewFile(packageName)
	f.HeaderComment("Code generated by generate-services. DO NOT EDIT.")

	// Option type
	f.Comment(fmt.Sprintf("Option defines a functional option for %s operations", packageName))
	f.Type().Id("Option").Func().Params(Op("*").Id("Options"))

	// Options struct
	var fields []Code
	for _, qp := range queryParams {
		fields = append(fields, Id(qp.GoName).Op("*").Id(qp.Type).Tag(map[string]string{"url": qp.Name + ",omitempty"}))
	}
	f.Comment(fmt.Sprintf("Options holds query parameters for %s operations", packageName))
	f.Type().Id("Options").Struct(fields...)

	// With* functions for each query param
	for _, qp := range queryParams {
		f.Comment(fmt.Sprintf("With%s sets the %s query parameter", qp.GoName, qp.Name))
		f.Func().Id("With" + qp.GoName).Params(Id(qp.Name).Id(qp.Type)).Id("Option").Block(
			Return(Func().Params(Id("o").Op("*").Id("Options")).Block(
				Id("o").Dot(qp.GoName).Op("=").Op("&").Id(qp.Name),
			)),
		)
	}

	// buildQueryString function
	var buildBody []Code
	buildBody = append(buildBody, Id("options").Op(":=").Op("&").Id("Options").Values())
	buildBody = append(buildBody, For(List(Id("_"), Id("opt")).Op(":=").Range().Id("opts")).Block(
		Id("opt").Call(Id("options")),
	))

	if len(queryParams) > 0 {
		buildBody = append(buildBody, Empty())
		buildBody = append(buildBody, Var().Id("params").Index().String())

		for _, qp := range queryParams {
			var formatStr string
			var formatVerb string
			switch qp.Type {
			case "string":
				formatVerb = "%s"
			case "int", "int64":
				formatVerb = "%d"
			case "bool":
				formatVerb = "%t"
			default:
				formatVerb = "%v"
			}
			formatStr = qp.Name + "=" + formatVerb

			var valueExpr Code
			if qp.Type == "string" {
				valueExpr = Qual("net/url", "QueryEscape").Call(Op("*").Id("options").Dot(qp.GoName))
			} else {
				valueExpr = Op("*").Id("options").Dot(qp.GoName)
			}

			buildBody = append(buildBody, If(Id("options").Dot(qp.GoName).Op("!=").Nil()).Block(
				Id("params").Op("=").Append(Id("params"), Qual("fmt", "Sprintf").Call(Lit(formatStr), valueExpr)),
			))
		}

		buildBody = append(buildBody, Empty())
		buildBody = append(buildBody, If(Len(Id("params")).Op("==").Lit(0)).Block(Return(Lit(""))))
		buildBody = append(buildBody, Return(Qual("strings", "Join").Call(Id("params"), Lit("&"))))
	} else {
		buildBody = append(buildBody, Return(Lit("")))
	}

	f.Comment("buildQueryString builds a query string from options")
	f.Func().Id("buildQueryString").Params(Id("opts").Op("...").Id("Option")).String().Block(buildBody...)

	// Write file
	outputPath := filepath.Join(packageDir, "options.go")
	return f.Save(outputPath)
}

// Utility functions
func toPascalCase(s string) string {
	if s == "" {
		return ""
	}

	// Split by common delimiters
	words := strings.FieldsFunc(s, func(c rune) bool {
		return c == '_' || c == '-' || c == ' '
	})

	var result strings.Builder
	for _, word := range words {
		if len(word) > 0 {
			result.WriteString(strings.ToUpper(string(word[0])))
			result.WriteString(strings.ToLower(word[1:]))
		}
	}

	// Apply ID suffix fix
	return fixIDSuffixes(result.String())
}

func toCamelCase(s string) string {
	pascal := toPascalCase(s)
	if len(pascal) == 0 {
		return ""
	}
	return strings.ToLower(string(pascal[0])) + pascal[1:]
}

func toSnakeCase(s string) string {
	if s == "" {
		return ""
	}

	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}

	return strings.ToLower(result.String())
}

func convertPathToGoFormat(path string) string {
	// Convert OpenAPI path parameters from {param} to %s for fmt.Sprintf
	result := path

	// Replace each {param} with %s
	for strings.Contains(result, "{") {
		start := strings.Index(result, "{")
		end := strings.Index(result, "}")
		if start != -1 && end != -1 && end > start {
			result = result[:start] + "%s" + result[end+1:]
		} else {
			break
		}
	}

	return result
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

func mapSchemaMapToGoType(schema map[string]any) string {
	if ref := getSchemaRef(schema); ref != "" {
		return "*models." + extractTypeFromRef(ref)
	}

	switch getSchemaType(schema) {
	case "string":
		return "string"
	case "integer":
		if getSchemaFormat(schema) == "int64" {
			return "int64"
		}
		return "int"
	case "number":
		if getSchemaFormat(schema) == "double" {
			return "float64"
		}
		return "float32"
	case "boolean":
		return "bool"
	case "array":
		if items, ok := schema["items"].(map[string]any); ok {
			itemType := mapSchemaMapToGoType(items)
			return "[]" + itemType
		}
		return "[]any"
	case "object":
		return "map[string]any"
	default:
		return "any"
	}
}

func extractTypeFromRef(ref string) string {
	// Extract type name from $ref like "#/components/schemas/Topic"
	parts := strings.Split(ref, "/")
	if len(parts) > 0 {
		typeName := parts[len(parts)-1]
		// Apply the same naming conventions as the model generator
		return toGoStructName(typeName)
	}
	return "any"
}

func toGoStructName(name string) string {
	// Preserve exact casing for certain names to avoid conflicts
	if name == "WireGuard" {
		return "WireGuard"
	}
	// Use the same PascalCase logic as the model generator
	return toPascalCaseForTypes(name)
}

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

func derivePackageFromPath(path string) string {
	// Extract package name from path like /v4/addon-providers/addon-pulsar/...
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for i, part := range parts {
		if part == "addon-providers" && i+1 < len(parts) {
			next := parts[i+1]
			if after, ok := strings.CutPrefix(next, "addon-"); ok {
				return after
			}
		}
	}

	// Fallback to first meaningful part
	if len(parts) > 1 {
		return toSnakeCase(parts[1])
	}

	return "unknown"
}

func fixIDSuffixes(s string) string {
	// Fix common Go naming conventions for ID-related suffixes
	// This ensures getPulsar becomes GetPulsar, getId becomes GetID, etc.
	replacements := map[string]string{
		"Id":   "ID",
		"Ip":   "IP",
		"Ipam": "IPAM",
		"Url":  "URL",
		"Uri":  "URI",
		"Api":  "API",
		"Tls":  "TLS",
		"Ttl":  "TTL",
		"Sql":  "SQL",
		"Http": "HTTP",
		"Https": "HTTPS",
		"Json": "JSON",
		"Xml":  "XML",
		"Uuid": "UUID",
	}

	for old, new := range replacements {
		// Replace at the end of the string (suffix)
		if strings.HasSuffix(s, old) {
			s = strings.TrimSuffix(s, old) + new
			continue
		}

		// Replace in the middle when followed by another capital letter (A-Z)
		for i := 'A'; i <= 'Z'; i++ {
			next := string(i)
			s = strings.ReplaceAll(s, old+next, new+next)
		}
	}

	return s
}

// getPackageNameFromService determines the package name for a given service name
// using SERVICE_NAME_EXCEPTIONS and falling back to snake_case conversion
func getPackageNameFromService(serviceName string) string {
	// Check SERVICE_NAME_EXCEPTIONS first
	if exceptionName, hasException := SERVICE_NAME_EXCEPTIONS[serviceName]; hasException {
		return exceptionName
	}
	// Fall back to snake_case conversion
	return toSnakeCase(serviceName)
}

func isReservedWord(word string) bool {
	reservedWords := map[string]bool{
		"break": true, "case": true, "chan": true, "const": true, "continue": true,
		"default": true, "defer": true, "else": true, "fallthrough": true, "for": true,
		"func": true, "go": true, "goto": true, "if": true, "import": true,
		"interface": true, "map": true, "package": true, "range": true, "return": true,
		"select": true, "struct": true, "switch": true, "type": true, "var": true,
	}
	return reservedWords[word]
}

// applyMappingRules applies the mapping rules from SERVICE_MAPPING_EXCEPTIONS.md
func applyMappingRules(operationID, path string, tags []string) string {
	// Rule 1: Toute opération avec "pulsar" dans le path → service pulsar
	if strings.Contains(strings.ToLower(path), "pulsar") {
		return "pulsar"
	}

	// Rule 2: Toute opération avec "Pulsar" dans l'operationId → service pulsar
	if strings.Contains(operationID, "Pulsar") || strings.Contains(strings.ToLower(operationID), "pulsar") {
		return "pulsar"
	}

	// Rule 3: Operations avec tag "pulsar" → service pulsar
	for _, tag := range tags {
		if strings.ToLower(tag) == "pulsar" {
			return "pulsar"
		}
	}

	// No mapping rule applies
	return ""
}

// fileExists checks if a file exists
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}
