// Package gotype turns the schema of a request body or a response into the Go
// type the SDK publishes for it, the same way for the services generator and
// the builder generator.
//
// The two used to derive these types separately, and only agreed on the shapes
// the OVD document had: a `$ref`, or an array of one. Everything else — a map of
// strings, an inline object, an array of strings — was silently dropped, so the
// function for `PUT /v2/organisations/{id}/applications/{appId}/env` took no
// body at all. The unified document carries every one of those shapes, and the
// Go type is now derived here, once.
package gotype

import (
	"sort"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/swaggest/openapi-go/openapi31"

	"go.clever-cloud.dev/sdk/internal/naming"
)

// ModelsPackage is the import path of the generated models.
const ModelsPackage = "go.clever-cloud.dev/sdk/models"

// Type is the Go type of a payload.
type Type struct {
	// Expr is the type as Go spells it, with `models.` in front of a component
	// type and the bare name of an inline struct: `*models.WannabeAddon`,
	// `[]models.AddonView`, `map[string]string`, `UpdateAddonInfoBody`,
	// `client.Nothing`. Empty when the operation has no such payload.
	Expr string
	// Inline lists the structs an inline object schema turns into, nested ones
	// first, for the services generator to emit next to the operation.
	Inline []Struct
}

// Struct is an inline object schema as a named Go struct.
type Struct struct {
	Name        string
	Description string
	Fields      []Field
}

// Field is one property of an inline object schema.
type Field struct {
	Name        string
	JSONName    string
	Expr        string
	Required    bool
	Description string
}

// None is the absence of a payload.
var None = Type{}

// Nothing is a response without a body.
var Nothing = Type{Expr: "client.Nothing"}

// Unknown is the response of an operation that documents no success response
// at all: whatever comes back is decoded as `any`, as the SDK has always done,
// so a caller can still look at it.
var Unknown = Type{Expr: "any"}

// IsNone reports whether the operation has no such payload.
func (t Type) IsNone() bool { return t.Expr == "" }

// IsNothing reports whether the response has no body.
func (t Type) IsNothing() bool { return t.Expr == Nothing.Expr }

// Body is the type of a request body. A component type is taken behind a
// pointer (`*models.X`, `[]*models.X`), as the SDK has always published it;
// an inline object becomes a struct named after the operation (`<Name>Body`).
func Body(content map[string]openapi31.MediaType, name string) Type {
	schema, mediaType, ok := pick(content)
	if !ok {
		return None
	}
	t := of(schema, name+"Body", true)
	if !isJSON(mediaType) && !strings.Contains(t.Expr, "models.") && len(t.Inline) == 0 {
		// A text or binary body the JSON client cannot send.
		return None
	}
	return t
}

// Response is the type of a response body. A component type is taken by value
// (`models.X`, `[]models.X`), as `client.Response[T]` has always carried it; an
// inline object becomes a struct named after the operation (`<Name>Response`).
// No content, or content the JSON client cannot decode, is Nothing.
func Response(content map[string]openapi31.MediaType, name string) Type {
	schema, mediaType, ok := pick(content)
	if !ok {
		return Nothing
	}
	t := of(schema, name+"Response", false)
	if !isJSON(mediaType) && !strings.Contains(t.Expr, "models.") && len(t.Inline) == 0 {
		return Nothing
	}
	return t
}

// pick chooses the media type a payload is generated from: `application/json`,
// then any JSON flavour (`application/json-patch+json`), then the first in
// alphabetical order — a fixed choice, where ranging over the map picked one at
// random on every run.
func pick(content map[string]openapi31.MediaType) (map[string]any, string, bool) {
	if len(content) == 0 {
		return nil, "", false
	}
	names := make([]string, 0, len(content))
	for name := range content {
		names = append(names, name)
	}
	sort.Strings(names)
	chosen := names[0]
	for _, name := range names {
		if name == "application/json" {
			chosen = name
			break
		}
		if isJSON(name) && !isJSON(chosen) {
			chosen = name
		}
	}
	return content[chosen].Schema, chosen, true
}

func isJSON(mediaType string) bool {
	return strings.Contains(mediaType, "json")
}

// of maps a schema to its Go type. refPointer puts component types behind a
// pointer; name is what an inline object struct is called.
func of(schema map[string]any, name string, refPointer bool) Type {
	if schema == nil {
		return Type{Expr: "any"}
	}
	if inner, ok := unwrapNullable(schema); ok {
		t := of(inner, name, refPointer)
		if strings.HasPrefix(t.Expr, "models.") || isPrimitive(t.Expr) {
			t.Expr = "*" + t.Expr
		}
		return t
	}
	if ref, ok := schema["$ref"].(string); ok {
		expr := "models." + naming.StructName(strings.TrimPrefix(ref, "#/components/schemas/"))
		if refPointer {
			expr = "*" + expr
		}
		return Type{Expr: expr}
	}
	types := schemaTypes(schema)
	nullable := false
	var kind string
	for _, t := range types {
		if t == "null" {
			nullable = true
		} else {
			kind = t
		}
	}
	var t Type
	switch kind {
	case "string":
		t = Type{Expr: "string"}
	case "integer":
		if schema["format"] == "int64" {
			t = Type{Expr: "int64"}
		} else {
			t = Type{Expr: "int"}
		}
	case "number":
		t = Type{Expr: "float64"}
	case "boolean":
		t = Type{Expr: "bool"}
	case "array":
		items, _ := schema["items"].(map[string]any)
		item := of(items, name+"Item", refPointer)
		t = Type{Expr: "[]" + item.Expr, Inline: item.Inline}
	case "object":
		if props, ok := schema["properties"].(map[string]any); ok && len(props) > 0 {
			return inlineStruct(schema, props, name)
		}
		if ap, ok := schema["additionalProperties"].(map[string]any); ok {
			value := of(ap, name+"Value", false)
			t = Type{Expr: "map[string]" + value.Expr, Inline: value.Inline}
		} else {
			t = Type{Expr: "map[string]any"}
		}
	default:
		// No type and no $ref: `{}`, a oneOf that is not the nullable idiom, an
		// allOf — anything goes.
		return Type{Expr: "any"}
	}
	if nullable && isPrimitive(t.Expr) {
		t.Expr = "*" + t.Expr
	}
	return t
}

// inlineStruct turns an object schema with properties into a named struct, its
// nested inline objects into structs of their own named after the field.
func inlineStruct(schema, props map[string]any, name string) Type {
	required := map[string]bool{}
	if list, ok := schema["required"].([]any); ok {
		for _, r := range list {
			if s, ok := r.(string); ok {
				required[s] = true
			}
		}
	}
	keys := make([]string, 0, len(props))
	for key := range props {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	st := Struct{Name: name}
	if desc, ok := schema["description"].(string); ok {
		st.Description = desc
	}
	var nested []Struct
	for _, key := range keys {
		prop, _ := props[key].(map[string]any)
		fieldName := naming.FieldName(key)
		ft := of(prop, name+fieldName, false)
		nested = append(nested, ft.Inline...)
		expr := ft.Expr
		if !required[key] && !strings.HasPrefix(expr, "*") && !strings.HasPrefix(expr, "[]") && !strings.HasPrefix(expr, "map[") && expr != "any" {
			expr = "*" + expr
		}
		field := Field{Name: fieldName, JSONName: key, Expr: expr, Required: required[key]}
		if desc, ok := prop["description"].(string); ok {
			field.Description = desc
		}
		st.Fields = append(st.Fields, field)
	}
	return Type{Expr: name, Inline: append(nested, st)}
}

func schemaTypes(schema map[string]any) []string {
	switch v := schema["type"].(type) {
	case string:
		return []string{v}
	case []any:
		var out []string
		for _, t := range v {
			if s, ok := t.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}

// unwrapNullable recognises OpenAPI 3.1's nullable idiom, `oneOf: [{type:
// null}, X]`, and returns X.
func unwrapNullable(schema map[string]any) (map[string]any, bool) {
	members, ok := schema["oneOf"].([]any)
	if !ok || len(members) != 2 {
		return nil, false
	}
	var inner map[string]any
	sawNull := false
	for _, m := range members {
		member, ok := m.(map[string]any)
		if !ok {
			return nil, false
		}
		if types := schemaTypes(member); len(types) == 1 && types[0] == "null" {
			sawNull = true
			continue
		}
		inner = member
	}
	if !sawNull || inner == nil {
		return nil, false
	}
	return inner, true
}

func isPrimitive(expr string) bool {
	switch expr {
	case "string", "int", "int64", "float64", "bool":
		return true
	}
	return false
}

// Code renders the type for jennifer. inline says how to refer to an inline
// struct: by name from inside the services package, qualified from outside.
func (t Type) Code(inline func(name string) jen.Code) jen.Code {
	return exprCode(t.Expr, inline)
}

func exprCode(expr string, inline func(name string) jen.Code) jen.Code {
	switch {
	case expr == "":
		return jen.Any()
	case expr == Nothing.Expr:
		return jen.Qual("go.clever-cloud.dev/client", "Nothing")
	case expr == "any":
		return jen.Any()
	case strings.HasPrefix(expr, "*"):
		return jen.Op("*").Add(exprCode(expr[1:], inline))
	case strings.HasPrefix(expr, "[]"):
		return jen.Index().Add(exprCode(expr[2:], inline))
	case strings.HasPrefix(expr, "map[string]"):
		return jen.Map(jen.String()).Add(exprCode(strings.TrimPrefix(expr, "map[string]"), inline))
	case strings.HasPrefix(expr, "models."):
		return jen.Qual(ModelsPackage, strings.TrimPrefix(expr, "models."))
	case expr == "string":
		return jen.String()
	case expr == "int":
		return jen.Int()
	case expr == "int64":
		return jen.Int64()
	case expr == "float64":
		return jen.Float64()
	case expr == "bool":
		return jen.Bool()
	default:
		return inline(expr)
	}
}

// Declare renders an inline struct as a type declaration.
func (s Struct) Declare() jen.Code {
	var fields []jen.Code
	for _, f := range s.Fields {
		tag := f.JSONName
		if !f.Required {
			tag += ",omitempty"
		}
		field := jen.Id(f.Name).Add(exprCode(f.Expr, func(name string) jen.Code { return jen.Id(name) })).Tag(map[string]string{"json": tag})
		if f.Description != "" {
			field = field.Comment(strings.ReplaceAll(strings.ReplaceAll(f.Description, "\n", " "), "*/", "* /"))
		}
		fields = append(fields, field)
	}
	return jen.Type().Id(s.Name).Struct(fields...)
}
