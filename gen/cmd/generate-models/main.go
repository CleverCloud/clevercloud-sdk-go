// Package main implements a code generator that creates Go models from OpenAPI specifications
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"maps"
	"os"
	"path/filepath"
	"sort"
	"strings"

	. "github.com/dave/jennifer/jen"
	"github.com/swaggest/openapi-go/openapi31"
	"gopkg.in/yaml.v3"
)

// Configuration for the generator
type Config struct {
	SpecFile    string
	OutputDir   string
	PackageName string
}

// ModelField represents a field in a generated model
type ModelField struct {
	Name       string
	Type       string
	JSONTag    string
	Comment    string
	IsRequired bool
	IsPointer  bool
	ConstValue string // If set, this field has a const value
}

// ModelStruct represents a Go struct to be generated
type ModelStruct struct {
	Name        string
	Comment     string
	Fields      []ModelField
	IsEnum      bool
	EnumValues  []string
	EnumType    string
	IsTypeAlias bool
	AliasType   string
	IsUnion     bool
	UnionTypes  []string
	TypeField   string   // Name of the const field that identifies the type (e.g., "Type")
	TypeValue   string   // Const value for this type (e.g., "DATADOG")
	Implements  []string // List of union interfaces this type implements
}

// PackageData represents the complete package to generate
type PackageData struct {
	Package string
	Imports []string
	Models  []ModelStruct
}

func main() {
	config := parseFlags()

	log.Printf("🚀 Starting model generation from %s", config.SpecFile)

	// Parse OpenAPI specification
	spec, err := parseOpenAPISpec(config.SpecFile)
	if err != nil {
		log.Fatalf("Failed to parse OpenAPI spec: %v", err)
	}

	// Generate models from schemas
	models, err := generateModels(spec)
	if err != nil {
		log.Fatalf("Failed to generate models: %v", err)
	}

	// Create package data
	packageData := &PackageData{
		Package: config.PackageName,
		Imports: []string{"time"},
		Models:  models,
	}

	// Generate Go files
	if err := generateGoFiles(packageData, config.OutputDir); err != nil {
		log.Fatalf("Failed to generate Go files: %v", err)
	}

	log.Printf("✅ Successfully generated %d models in %s", len(models), config.OutputDir)
}

func parseFlags() *Config {
	config := &Config{}

	flag.StringVar(&config.SpecFile, "spec", "openapi.json", "Path to OpenAPI specification file")
	flag.StringVar(&config.OutputDir, "output", "models", "Output directory for generated models")
	flag.StringVar(&config.PackageName, "package", "models", "Go package name")
	flag.Parse()

	return config
}

// Schema is an alias for map[string]any used by openapi31
type Schema = map[string]any

// Helper functions to extract values from Schema (map[string]any)
func getSchemaType(schema Schema) []string {
	if t, ok := schema["type"].(string); ok {
		return []string{t}
	}
	if arr, ok := schema["type"].([]any); ok {
		result := make([]string, len(arr))
		for i, v := range arr {
			if s, ok := v.(string); ok {
				result[i] = s
			}
		}
		return result
	}
	return nil
}

func getSchemaTitle(schema Schema) string {
	if t, ok := schema["title"].(string); ok {
		return t
	}
	return ""
}

func getSchemaDescription(schema Schema) string {
	if d, ok := schema["description"].(string); ok {
		return d
	}
	return ""
}

func getSchemaRequired(schema Schema) []string {
	if arr, ok := schema["required"].([]any); ok {
		result := make([]string, len(arr))
		for i, v := range arr {
			if s, ok := v.(string); ok {
				result[i] = s
			}
		}
		return result
	}
	return nil
}

func getSchemaProperties(schema Schema) map[string]Schema {
	if props, ok := schema["properties"].(map[string]any); ok {
		result := make(map[string]Schema)
		for k, v := range props {
			if m, ok := v.(map[string]any); ok {
				result[k] = m
			}
		}
		return result
	}
	return nil
}

func getSchemaEnum(schema Schema) []string {
	if arr, ok := schema["enum"].([]any); ok {
		result := make([]string, len(arr))
		for i, v := range arr {
			if s, ok := v.(string); ok {
				result[i] = s
			} else {
				result[i] = fmt.Sprintf("%v", v)
			}
		}
		return result
	}
	return nil
}

func getSchemaItems(schema Schema) Schema {
	if items, ok := schema["items"].(map[string]any); ok {
		return items
	}
	return nil
}

func getSchemaFormat(schema Schema) string {
	if f, ok := schema["format"].(string); ok {
		return f
	}
	return ""
}

func getSchemaRef(schema Schema) string {
	if ref, ok := schema["$ref"].(string); ok {
		return ref
	}
	return ""
}

func getSchemaOneOf(schema Schema) []Schema {
	if arr, ok := schema["oneOf"].([]any); ok {
		result := make([]Schema, len(arr))
		for i, v := range arr {
			if m, ok := v.(map[string]any); ok {
				result[i] = m
			}
		}
		return result
	}
	return nil
}

func getSchemaConst(schema Schema) (string, bool) {
	if c, ok := schema["const"].(string); ok {
		return c, true
	}
	return "", false
}

func getSchemaAdditionalProperties(schema Schema) (Schema, bool) {
	if ap, ok := schema["additionalProperties"]; ok {
		// Can be true (any type) or a schema object
		if apMap, ok := ap.(map[string]any); ok {
			return apMap, true
		}
		if apBool, ok := ap.(bool); ok && apBool {
			// additionalProperties: true means map[string]any
			return nil, true
		}
	}
	return nil, false
}

func parseOpenAPISpec(specFile string) (*openapi31.Spec, error) {
	// Read OpenAPI spec file
	specBytes, err := os.ReadFile(specFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read spec file: %w", err)
	}

	var spec openapi31.Spec

	// Detect format based on file extension
	ext := strings.ToLower(filepath.Ext(specFile))

	switch ext {
	case ".json":
		// Parse JSON document
		if err := json.Unmarshal(specBytes, &spec); err != nil {
			return nil, fmt.Errorf("failed to parse OpenAPI JSON: %w", err)
		}
	case ".yaml", ".yml":
		// Parse YAML document
		if err := yaml.Unmarshal(specBytes, &spec); err != nil {
			return nil, fmt.Errorf("failed to parse OpenAPI YAML: %w", err)
		}
	default:
		// Try to auto-detect format
		log.Printf("Unknown file extension %s, attempting auto-detection...", ext)

		// Try JSON first
		err := json.Unmarshal(specBytes, &spec)
		if err == nil {
			log.Printf("Auto-detected JSON format")
		} else {
			// Try YAML
			if err := yaml.Unmarshal(specBytes, &spec); err != nil {
				return nil, fmt.Errorf("failed to parse OpenAPI file (tried both JSON and YAML): %w", err)
			}
			log.Printf("Auto-detected YAML format")
		}
	}

	return &spec, nil
}

func generateModels(spec *openapi31.Spec) ([]ModelStruct, error) {
	var models []ModelStruct

	if spec.Components == nil || spec.Components.Schemas == nil {
		return models, nil
	}

	// Sort schema names alphabetically for consistent output
	var schemaNames []string
	for schemaName := range spec.Components.Schemas {
		schemaNames = append(schemaNames, schemaName)
	}
	sort.Strings(schemaNames)

	// Process each schema in components/schemas
	for _, schemaName := range schemaNames {
		schema := spec.Components.Schemas[schemaName]
		model, err := processSchema(schemaName, schema)
		if err != nil {
			log.Printf("Warning: Failed to process schema %s: %v", schemaName, err)
			continue
		}

		if model != nil {
			models = append(models, *model)
		}
	}

	// Sort models by name for consistent output
	sort.Slice(models, func(i, j int) bool {
		return models[i].Name < models[j].Name
	})

	// Build map of unions and their member types
	unionMembers := make(map[string][]string)
	for _, m := range models {
		if m.IsUnion {
			unionMembers[m.Name] = m.UnionTypes
		}
	}

	// Associate types with unions they implement
	for i := range models {
		for unionName, members := range unionMembers {
			for _, member := range members {
				if models[i].Name == member {
					models[i].Implements = append(models[i].Implements, unionName)
				}
			}
		}
	}

	return models, nil
}

func processSchema(name string, schema Schema) (*ModelStruct, error) {
	// Handle enum types
	if len(getSchemaEnum(schema)) > 0 {
		return processEnumSchema(name, schema)
	}

	// Handle oneOf union types
	if len(getSchemaOneOf(schema)) > 0 {
		return processOneOfSchema(name, schema)
	}

	// Handle object types
	schemaTypes := getSchemaType(schema)
	if len(schemaTypes) > 0 && schemaTypes[0] == "object" {
		// Check if this is a map type (object with additionalProperties)
		if apSchema, hasAP := getSchemaAdditionalProperties(schema); hasAP {
			// Check if there are also regular properties
			properties := getSchemaProperties(schema)
			if len(properties) == 0 {
				// Pure map type (no fixed properties, only additionalProperties)
				return processMapSchema(name, schema, apSchema)
			}
			// Mixed object with both properties and additionalProperties
			// Treat as regular object for now
		}
		return processObjectSchema(name, schema)
	}

	// Handle scalar types with titles (type aliases)
	if len(schemaTypes) > 0 && getSchemaTitle(schema) != "" {
		return processTypeAliasSchema(name, schema)
	}

	// Skip other schemas
	return nil, nil
}

func processMapSchema(name string, schema Schema, additionalPropertiesSchema Schema) (*ModelStruct, error) {
	var valueType string

	if additionalPropertiesSchema == nil {
		// additionalProperties: true means map[string]any
		valueType = "any"
	} else {
		// Get the type of the values
		vt, _, err := getGoType(additionalPropertiesSchema, true)
		if err != nil {
			return nil, err
		}
		valueType = vt
	}

	// Create a type alias for the map
	mapType := fmt.Sprintf("map[string]%s", valueType)
	model := &ModelStruct{
		Name:        toGoStructName(name),
		Comment:     formatComment(getSchemaDescription(schema)),
		IsTypeAlias: true,
		AliasType:   mapType,
	}

	return model, nil
}

func processEnumSchema(name string, schema Schema) (*ModelStruct, error) {
	// Determine the enum's base type from the schema
	enumType := "string" // Default to string
	schemaTypes := getSchemaType(schema)
	if len(schemaTypes) > 0 {
		switch schemaTypes[0] {
		case "integer":
			enumType = "int"
		case "number":
			enumType = "float64"
		case "boolean":
			enumType = "bool"
		case "string":
			enumType = "string"
		}
	}

	model := &ModelStruct{
		Name:       toGoStructName(name),
		Comment:    formatComment(getSchemaDescription(schema)),
		IsEnum:     true,
		EnumType:   enumType,
		EnumValues: getSchemaEnum(schema),
	}

	return model, nil
}

func processTypeAliasSchema(name string, schema Schema) (*ModelStruct, error) {
	// For type aliases, we need to get the underlying primitive type
	// not the title, so we create a copy without the title
	schemaForType := make(Schema)
	maps.Copy(schemaForType, schema)
	delete(schemaForType, "title")

	// Get the underlying Go type
	goType, _, err := getGoType(schemaForType, true)
	if err != nil {
		return nil, err
	}

	// Create a type alias
	model := &ModelStruct{
		Name:        toGoStructName(name),
		Comment:     formatComment(getSchemaDescription(schema)),
		IsTypeAlias: true,
		AliasType:   goType,
	}

	return model, nil
}

func processOneOfSchema(name string, schema Schema) (*ModelStruct, error) {
	// For oneOf, we'll create an interface type in Go
	var unionTypes []string

	for _, oneOfSchema := range getSchemaOneOf(schema) {
		if ref := getSchemaRef(oneOfSchema); ref != "" {
			// Extract type name from $ref
			refName := strings.TrimPrefix(ref, "#/components/schemas/")
			unionTypes = append(unionTypes, toGoStructName(refName))
		} else if len(getSchemaType(oneOfSchema)) > 0 {
			// Get the Go type for primitive types
			goType, _, err := getGoType(oneOfSchema, true)
			if err != nil {
				return nil, err
			}
			unionTypes = append(unionTypes, goType)
		}
	}

	model := &ModelStruct{
		Name:       toGoStructName(name),
		Comment:    formatComment(getSchemaDescription(schema)),
		IsUnion:    true,
		UnionTypes: unionTypes,
	}

	return model, nil
}

func processObjectSchema(name string, schema Schema) (*ModelStruct, error) {
	model := &ModelStruct{
		Name:    toGoStructName(name),
		Comment: formatComment(getSchemaDescription(schema)),
		Fields:  make([]ModelField, 0),
	}

	// Get required fields
	requiredFields := make(map[string]bool)
	for _, req := range getSchemaRequired(schema) {
		requiredFields[req] = true
	}

	// Sort property names alphabetically for consistent output
	properties := getSchemaProperties(schema)
	var propNames []string
	for propName := range properties {
		propNames = append(propNames, propName)
	}
	sort.Strings(propNames)

	// Process properties
	for _, propName := range propNames {
		propSchema := properties[propName]
		field, err := processProperty(propName, propSchema, requiredFields[propName])
		if err != nil {
			log.Printf("Warning: Failed to process property %s.%s: %v", name, propName, err)
			continue
		}

		// If this field has a const value, track it
		if field.ConstValue != "" {
			model.TypeField = field.Name
			model.TypeValue = field.ConstValue
		}

		model.Fields = append(model.Fields, *field)
	}

	// Sort fields by name for consistent output
	sort.Slice(model.Fields, func(i, j int) bool {
		return model.Fields[i].Name < model.Fields[j].Name
	})

	return model, nil
}

func processProperty(propName string, propSchema Schema, isRequired bool) (*ModelField, error) {
	field := &ModelField{
		Name:       toGoFieldName(propName),
		JSONTag:    buildJSONTag(propName, isRequired),
		Comment:    formatComment(getSchemaDescription(propSchema)),
		IsRequired: isRequired,
	}

	// Check for const value
	if constVal, hasConst := getSchemaConst(propSchema); hasConst {
		field.ConstValue = constVal
	}

	// Determine Go type
	goType, isPointer, err := getGoType(propSchema, isRequired)
	if err != nil {
		return nil, err
	}

	field.Type = goType
	field.IsPointer = isPointer

	return field, nil
}

func getGoType(schema Schema, isRequired bool) (string, bool, error) {
	// Handle $ref types
	if ref := getSchemaRef(schema); ref != "" {
		refName := strings.TrimPrefix(ref, "#/components/schemas/")
		return toGoStructName(refName), !isRequired, nil
	}

	schemaTypes := getSchemaType(schema)
	if len(schemaTypes) == 0 {
		// Check title for enum references
		if title := getSchemaTitle(schema); title != "" {
			return toGoStructName(title), !isRequired, nil
		}
		return "any", !isRequired, nil
	}

	schemaType := schemaTypes[0]
	isPointer := !isRequired

	switch schemaType {
	case "string":
		if getSchemaFormat(schema) == "date-time" {
			return "time.Time", isPointer, nil
		}
		// Check if this is an enum reference
		if title := getSchemaTitle(schema); title != "" {
			return toGoStructName(title), isPointer, nil
		}
		return "string", isPointer, nil

	case "integer":
		return "int", isPointer, nil

	case "number":
		return "float64", isPointer, nil

	case "boolean":
		return "bool", isPointer, nil

	case "array":
		if items := getSchemaItems(schema); items != nil {
			itemType, _, err := getGoType(items, true)
			if err != nil {
				return "", false, err
			}
			// For slices, never use pointer even if optional - empty slice is the zero value
			return "[]" + itemType, false, nil
		}
		// For slices, never use pointer even if optional - empty slice is the zero value
		return "[]any", false, nil

	case "object":
		return "map[string]any", isPointer, nil

	default:
		return "any", isPointer, nil
	}
}

func buildJSONTag(propName string, isRequired bool) string {
	if isRequired {
		return fmt.Sprintf("`json:\"%s\"`", propName)
	}
	return fmt.Sprintf("`json:\"%s,omitempty\"`", propName)
}

func toGoStructName(name string) string {
	// Preserve exact casing for certain names to avoid conflicts
	if name == "WireGuard" {
		return "WireGuard"
	}
	return toPascalCase(name)
}

func toGoFieldName(name string) string {
	// Handle special cases for better Go naming
	switch strings.ToLower(name) {
	case "id":
		return "ID"
	case "url":
		return "URL"
	case "uri":
		return "URI"
	case "api":
		return "API"
	case "ttl":
		return "TTL"
	}

	// Convert snake_case and kebab-case to PascalCase
	result := toPascalCase(name)

	// Fix common Go naming conventions for ID suffixes
	result = fixIdSuffixes(result)

	return result
}

func fixIdSuffixes(s string) string {
	// Fix common Go naming conventions for ID-related suffixes
	// This ensures clusterId becomes ClusterID, ownerId becomes OwnerID, etc.
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

func toPascalCase(s string) string {
	// Handle special cases and clean input
	s = strings.ReplaceAll(s, ".", "_DOT_")
	s = strings.ReplaceAll(s, "+", "_PLUS_")
	s = strings.ReplaceAll(s, "*", "_STAR_")
	s = strings.ReplaceAll(s, "/", "_SLASH_")

	// Check if string contains separators (underscore, hyphen, space)
	hasSeparators := strings.ContainsAny(s, "_- ")

	if !hasSeparators {
		// No separators - preserve original casing (already PascalCase or camelCase)
		// Just ensure first letter is uppercase, keep rest as-is
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

func formatComment(description string) string {
	if description == "" {
		return ""
	}

	// Clean up the description
	desc := strings.TrimSpace(description)
	desc = strings.ReplaceAll(desc, "\n", " ")

	// Limit length
	if len(desc) > 100 {
		desc = desc[:97] + "..."
	}

	return desc
}

func generateFileName(typeName, suffix string) string {
	// Convert to snake_case to avoid file name conflicts with case-insensitive file systems
	// but preserve unique identifiers
	return toSnakeCase(typeName) + suffix
}

func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteByte('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

func generateGoFiles(packageData *PackageData, outputDir string) error {
	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate individual files for each model
	enumCount := 0
	structCount := 0
	aliasCount := 0

	for _, model := range packageData.Models {
		if model.IsEnum {
			if err := generateSingleEnumFile(model, packageData, outputDir); err != nil {
				log.Printf("Warning: Failed to generate enum file for %s: %v", model.Name, err)
				continue
			}
			enumCount++
		} else if model.IsTypeAlias {
			if err := generateSingleTypeAliasFile(model, packageData, outputDir); err != nil {
				log.Printf("Warning: Failed to generate type alias file for %s: %v", model.Name, err)
				continue
			}
			aliasCount++
		} else if model.IsUnion {
			if err := generateSingleUnionFile(model, packageData, outputDir); err != nil {
				log.Printf("Warning: Failed to generate union file for %s: %v", model.Name, err)
				continue
			}
			aliasCount++ // Count unions as aliases for simplicity
		} else {
			if err := generateSingleStructFile(model, packageData, outputDir); err != nil {
				log.Printf("Warning: Failed to generate struct file for %s: %v", model.Name, err)
				continue
			}
			structCount++
		}
	}

	log.Printf("📊 Generated %d enum files, %d struct files, and %d type alias files", enumCount, structCount, aliasCount)
	return nil
}

func generateSingleEnumFile(enum ModelStruct, packageData *PackageData, outputDir string) error {
	f := NewFile(packageData.Package)
	f.HeaderComment("Code generated by generate-models. DO NOT EDIT.")

	// Type declaration
	f.Comment(fmt.Sprintf("%s represents %s", enum.Name, enum.Comment))
	f.Type().Id(enum.Name).Id(enum.EnumType)

	// Const block
	var defs []Code
	for _, value := range enum.EnumValues {
		defs = append(defs, Id(enum.Name+toPascalCase(value)).Id(enum.Name).Op("=").Lit(value))
	}
	f.Const().Defs(defs...)
	f.Line()

	// Generate method to convert to native type
	methodName, returnType := getEnumMethodName(enum.EnumType)
	f.Comment(fmt.Sprintf("%s returns the underlying %s value", methodName, enum.EnumType))
	f.Func().Params(Id("e").Id(enum.Name)).Id(methodName).Params().Id(returnType).Block(
		Return(Id(returnType).Parens(Id("e"))),
	)

	// Write to individual file
	fileName := generateFileName(enum.Name, "_enum.go")
	outputFile := filepath.Join(outputDir, fileName)
	return f.Save(outputFile)
}

// getEnumMethodName returns the method name and return type for an enum's native type conversion
func getEnumMethodName(enumType string) (methodName string, returnType string) {
	switch enumType {
	case "string":
		return "String", "string"
	case "int":
		return "Int", "int"
	case "int32":
		return "Int32", "int32"
	case "int64":
		return "Int64", "int64"
	case "float32":
		return "Float32", "float32"
	case "float64":
		return "Float64", "float64"
	case "bool":
		return "Bool", "bool"
	default:
		// Default to String for unknown types
		return "String", enumType
	}
}

func generateSingleTypeAliasFile(alias ModelStruct, packageData *PackageData, outputDir string) error {
	f := NewFile(packageData.Package)
	f.HeaderComment("Code generated by generate-models. DO NOT EDIT.")

	// Type declaration with comment
	f.Comment(fmt.Sprintf("%s %s", alias.Name, alias.Comment))
	f.Type().Id(alias.Name).Op("=").Add(parseTypeCode(alias.AliasType))

	// Write to individual file
	fileName := generateFileName(alias.Name, "_alias.go")
	outputFile := filepath.Join(outputDir, fileName)
	return f.Save(outputFile)
}

func generateSingleUnionFile(union ModelStruct, packageData *PackageData, outputDir string) error {
	f := NewFile(packageData.Package)
	f.HeaderComment("Code generated by generate-models. DO NOT EDIT.")

	// Comment with union types
	f.Comment(fmt.Sprintf("%s %s", union.Name, union.Comment))
	f.Comment(fmt.Sprintf("Union type - can be one of: %s", strings.Join(union.UnionTypes, ", ")))

	// Interface with marker method and GetType()
	f.Type().Id(union.Name).Interface(
		Id("is" + union.Name).Params(),
		Id("GetType").Params().String(),
	)

	// Write to individual file
	fileName := generateFileName(union.Name, "_union.go")
	outputFile := filepath.Join(outputDir, fileName)
	return f.Save(outputFile)
}

func generateSingleStructFile(model ModelStruct, packageData *PackageData, outputDir string) error {
	f := NewFile(packageData.Package)
	f.HeaderComment("Code generated by generate-models. DO NOT EDIT.")

	// Generate const for the type value if present
	if model.TypeValue != "" && model.TypeField != "" {
		constName := model.Name + model.TypeField
		f.Const().Id(constName).Op("=").Lit(model.TypeValue)
		f.Line()
	}

	// Struct comment
	f.Comment(fmt.Sprintf("%s %s", model.Name, model.Comment))

	// Build struct fields
	var fields []Code
	for _, field := range model.Fields {
		fieldCode := Id(field.Name)
		if field.IsPointer {
			fieldCode = fieldCode.Op("*")
		}
		fieldCode = fieldCode.Add(parseTypeCode(field.Type))

		// Parse JSON tag
		if field.JSONTag != "" {
			// Extract the tag value from `json:"..."` format
			tagValue := strings.TrimPrefix(field.JSONTag, "`json:\"")
			tagValue = strings.TrimSuffix(tagValue, "\"`")
			fieldCode = fieldCode.Tag(map[string]string{"json": tagValue})
		}

		if field.Comment != "" {
			fieldCode = fieldCode.Comment(field.Comment)
		}
		fields = append(fields, fieldCode)
	}

	f.Type().Id(model.Name).Struct(fields...)
	f.Line()

	// Generate GetType() method if this struct has a type const
	if model.TypeValue != "" && model.TypeField != "" {
		constName := model.Name + model.TypeField
		f.Comment(fmt.Sprintf("GetType returns the type identifier for %s", model.Name))
		f.Func().Params(Id("r").Id(model.Name)).Id("GetType").Params().String().Block(
			Return(Id(constName)),
		)
		f.Line()
	}

	// Generate marker methods for union interfaces this type implements
	for _, unionName := range model.Implements {
		markerMethod := "is" + unionName
		f.Comment(fmt.Sprintf("%s implements %s", markerMethod, unionName))
		f.Func().Params(Id("r").Id(model.Name)).Id(markerMethod).Params().Block()
		f.Line()
	}

	// Write to individual file
	fileName := generateFileName(model.Name, ".go")
	outputFile := filepath.Join(outputDir, fileName)
	return f.Save(outputFile)
}

// parseTypeCode converts a type string to jennifer Code, handling time.Time imports
func parseTypeCode(typeStr string) Code {
	if typeStr == "time.Time" {
		return Qual("time", "Time")
	}
	if strings.HasPrefix(typeStr, "[]") {
		inner := strings.TrimPrefix(typeStr, "[]")
		return Index().Add(parseTypeCode(inner))
	}
	if strings.HasPrefix(typeStr, "*") {
		inner := strings.TrimPrefix(typeStr, "*")
		return Op("*").Add(parseTypeCode(inner))
	}
	if strings.HasPrefix(typeStr, "map[") {
		// Simple map handling: map[string]any
		if typeStr == "map[string]any" {
			return Map(String()).Any()
		}
	}
	return Id(typeStr)
}
