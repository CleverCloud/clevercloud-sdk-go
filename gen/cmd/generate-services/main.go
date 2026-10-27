package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	. "github.com/dave/jennifer/jen"
	"github.com/swaggest/openapi-go/openapi31"

	"go.clever-cloud.dev/sdk/internal/gotype"
	"go.clever-cloud.dev/sdk/internal/naming"
	"go.clever-cloud.dev/sdk/internal/spec"
)

// Package names, function names and the exception tables live in internal/naming; the
// document is pruned and normalised by internal/spec. Both are shared with the builder
// generator so the two cannot drift apart.

// Service generation structures
type ServiceOperation struct {
	OperationID    string
	FunctionName   string
	PackageName    string
	Path           string
	Method         string
	Summary        string
	Description    string
	PathParams     []ServiceParam
	QueryParams    []ServiceParam
	Body           gotype.Type
	Response       gotype.Type
	XService       string
	TraceSpanName  string
	HasRequestBody bool
	HasQueryParams bool
}

type ServiceParam struct {
	// Name is the Go identifier: the spec's name, made legal (`for` is `forParam`,
	// `multi-slots` is `multi_slots`).
	Name string
	// WireName is the name the API expects, exactly as the spec spells it. A query
	// string built from Name sent `forParam=` where the API reads `for=`.
	WireName    string
	GoName      string
	Type        string
	Required    bool
	Description string
	IsPath      bool
	IsQuery     bool
}

func main() {
	var (
		specFile  = flag.String("spec", "openapi.json", "Path to OpenAPI specification file")
		outputDir = flag.String("output", "services", "Output directory for generated services")
	)
	flag.Parse()

	if *specFile == "" {
		log.Fatal("Please provide -spec flag with path to OpenAPI specification")
	}

	document, report, err := spec.Load(*specFile)
	if err != nil {
		log.Fatalf("Failed to load spec: %v", err)
	}
	log.Printf("Loaded %s: %s", *specFile, report)

	// Process operations
	operations := extractOperations(*document)

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

		// Ids are unique and x-service is set on every operation: spec.Load saw to it.
		operationID := ""
		if operation.ID != nil {
			operationID = *operation.ID
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
			FunctionName:  naming.FunctionName(operationID),
			Path:          convertPathToGoFormat(path),
			Method:        method,
			Summary:       summary,
			Description:   description,
			TraceSpanName: operationID,
		}

		xService, _ := operation.MapOfAnything["x-service"].(string)
		op.XService = xService
		op.PackageName = naming.OperationPackage(operationID, xService, path, operation.Tags)
		if op.PackageName == "" {
			log.Fatalf("%s %s (%s): no x-service and no tag to name a package after", method, path, operationID)
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
				WireName:    param.Name,
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

		// Request body and response: one derivation shared with the builder generator.
		if operation.RequestBody != nil && operation.RequestBody.RequestBody != nil {
			op.Body = gotype.Body(operation.RequestBody.RequestBody.Content, op.FunctionName)
			op.HasRequestBody = !op.Body.IsNone()
		}
		op.Response = gotype.Unknown
		if operation.Responses != nil {
			for _, statusCode := range []string{"200", "201", "202", "204"} {
				if respOrRef, exists := operation.Responses.MapOfResponseOrReferenceValues[statusCode]; exists {
					if resp := respOrRef.Response; resp != nil {
						op.Response = gotype.Response(resp.Content, op.FunctionName)
					}
					break
				}
			}
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

	// Generate one file per operation. Two ids can share a file name — `deleteCephPool` and
	// `delete_ceph_pool` are distinct functions on distinct routes but both spell
	// `delete_ceph_pool.go` — so the second takes its method as a suffix rather than
	// overwriting the first.
	fileNames := map[string]bool{}
	for _, op := range operations {
		fileName := operationFileName(op.OperationID)
		if fileNames[fileName] {
			fileName = strings.TrimSuffix(fileName, ".go") + "_" + strings.ToLower(op.Method) + ".go"
		}
		fileNames[fileName] = true
		if err := generateOperationFile(packageDir, fileName, op); err != nil {
			return fmt.Errorf("failed to generate operation %s: %v", op.OperationID, err)
		}
	}

	// Generate options.go file for the package
	if err := generateOptionsFile(packageDir, packageName, operations); err != nil {
		return fmt.Errorf("failed to generate options.go: %v", err)
	}

	return nil
}

func generateOperationFile(packageDir, fileName string, op ServiceOperation) error {
	f := NewFile(op.PackageName)
	f.HeaderComment("Code generated by generate-services. DO NOT EDIT.")

	inline := func(name string) Code { return Id(name) }
	responseType := op.Response.Code(inline)

	// Build function comment
	commentLines := []string{op.FunctionName + " " + sanitizeComment(op.Summary)}
	if op.Description != "" {
		commentLines = append(commentLines, "", sanitizeComment(op.Description))
	}
	commentLines = append(commentLines, "", "Parameters:", "  - ctx: context for the request", "  - client: the Clever Cloud client", "  - tracer: OpenTelemetry tracer for observability")
	for _, p := range op.PathParams {
		commentLines = append(commentLines, fmt.Sprintf("  - %s: %s", p.Name, sanitizeComment(p.Description)))
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
		params = append(params, Id("requestBody").Add(op.Body.Code(inline)))
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
	// An inline object schema has no component to refer to: its struct is declared here,
	// named after the operation, and the builder refers to it through this package.
	for _, st := range append(append([]gotype.Struct{}, op.Body.Inline...), op.Response.Inline...) {
		if st.Description != "" {
			f.Comment(st.Name + " " + sanitizeComment(st.Description))
		}
		f.Add(st.Declare())
		f.Line()
	}

	f.Comment(strings.Join(commentLines, "\n"))
	f.Func().Id(op.FunctionName).Params(params...).Qual("go.clever-cloud.dev/client", "Response").Types(responseType).Block(body...)

	// Write file
	filename := fileName
	outputPath := filepath.Join(packageDir, filename)
	return f.Save(outputPath)
}

// formatResponseTypeJen returns a jennifer Code for the response type
func generateOptionsFile(packageDir, packageName string, operations []ServiceOperation) error {
	// Collect unique query params from all operations in this package
	queryParamMap := make(map[string]ServiceParam)
	for _, op := range operations {
		for _, qp := range op.QueryParams {
			// Use the original name as key to avoid duplicates
			if _, exists := queryParamMap[qp.WireName]; !exists {
				// For options, use PascalCase for GoName and simplify complex types to string
				param := qp
				param.GoName = toPascalCase(qp.Name)
				// Query params should be simple types - convert complex refs to string,
				// and a list of them (`status=[]*models.ClusterStatusType`) to a list of strings.
				if strings.Contains(param.Type, "models.") {
					if strings.HasPrefix(param.Type, "[]") {
						param.Type = "[]string"
					} else {
						param.Type = "string"
					}
				}
				queryParamMap[qp.WireName] = param
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
		fields = append(fields, Id(qp.GoName).Op("*").Id(qp.Type).Tag(map[string]string{"url": qp.WireName + ",omitempty"}))
	}
	f.Comment(fmt.Sprintf("Options holds query parameters for %s operations", packageName))
	f.Type().Id("Options").Struct(fields...)

	// With* functions for each query param
	for _, qp := range queryParams {
		f.Comment(fmt.Sprintf("With%s sets the %s query parameter", qp.GoName, qp.WireName))
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
			formatStr = qp.WireName + "=" + formatVerb

			var valueExpr Code
			if qp.Type == "string" {
				valueExpr = Qual("net/url", "QueryEscape").Call(Op("*").Id("options").Dot(qp.GoName))
			} else {
				valueExpr = Op("*").Id("options").Dot(qp.GoName)
			}

			if strings.HasPrefix(qp.Type, "[]") {
				// OpenAPI's default for a query array is `form` + `explode`: the key once per
				// value (`status=A&status=B`). `%v` of the slice sent `status=[A B]`.
				buildBody = append(buildBody, If(Id("options").Dot(qp.GoName).Op("!=").Nil()).Block(
					For(List(Id("_"), Id("v")).Op(":=").Range().Op("*").Id("options").Dot(qp.GoName)).Block(
						Id("params").Op("=").Append(Id("params"), Qual("fmt", "Sprintf").Call(Lit(qp.WireName+"=%s"),
							Qual("net/url", "QueryEscape").Call(Qual("fmt", "Sprint").Call(Id("v"))))),
					),
				))
				continue
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
	return naming.PascalCase(s)
}

func toCamelCase(s string) string {
	pascal := toPascalCase(s)
	if len(pascal) == 0 {
		return ""
	}
	return strings.ToLower(string(pascal[0])) + pascal[1:]
}

// operationFileName is the file an operation is written to: its id with every capital
// turned into an underscore and a lower-case letter. Only a file name, not an identifier,
// so acronyms are not kept together (`getAIEndpoint` is `get_a_i_endpoint.go`).
func operationFileName(operationID string) string {
	var result strings.Builder
	for i, r := range operationID {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String()) + ".go"
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
	return naming.StructName(name)
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

// sanitizeComment makes spec prose safe to embed in a Go block comment.
//
// Descriptions in the OpenAPI document carry Markdown prose: MIME wildcards like `*/*`, and globs
// like `KV_*/REDIS_*`. Each of those contains `*/`, which closes the generated block comment early
// and leaves the remainder of the description as bare tokens — "expected declaration, found '*'".
//
// The sequence is spaced rather than dropped so the prose still reads correctly.
func sanitizeComment(text string) string {
	return strings.ReplaceAll(text, "*/", "* /")
}
