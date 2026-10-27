// Package spec loads the OpenAPI document the SDK is generated from and
// normalises it into what the generators can consume: one operation per route,
// a unique operation id on each, a service on each, every `$ref` resolved.
//
// The public document already satisfies all of that. The rules are kept as a
// guard: a route published under two spellings of a path parameter, an operation
// id used twice or a `$ref` to a missing schema would otherwise make each
// generator fail in its own way. Every rule lives here, applied once, and each
// decision on a named operation is logged.
package spec

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/swaggest/openapi-go/openapi31"
	"gopkg.in/yaml.v3"

	"go.clever-cloud.dev/sdk/internal/naming"
)

// Methods is the order operations are visited in, so every run and every
// generator sees them in the same sequence.
var Methods = []string{"DELETE", "GET", "PATCH", "POST", "PUT"}

// Report counts what Normalize changed, for the generator to print.
type Report struct {
	// Operations the document carried before normalisation.
	Operations int
	// Operations dropped because they reference a schema the document does not define.
	Dangling int
	// Operations dropped as the other spelling of a route already kept.
	Twins int
	// Operations given a mechanical id because theirs was taken.
	Renamed int
	// Operations given an `x-service` from their tags or path.
	Tagged int
	// Path parameters declared on behalf of an operation whose template names one it does not.
	Declared int
	// Operations left after normalisation.
	Kept int
}

// String is the one-line summary a generator logs.
func (r Report) String() string {
	return fmt.Sprintf("%d operations: kept %d, dropped %d with dangling $ref and %d twin spellings, renamed %d duplicate ids, derived x-service for %d, declared %d missing path parameters",
		r.Operations, r.Kept, r.Dangling, r.Twins, r.Renamed, r.Tagged, r.Declared)
}

// Load reads the document at path and normalises it.
func Load(path string) (*openapi31.Spec, Report, error) {
	spec, err := Parse(path)
	if err != nil {
		return nil, Report{}, err
	}
	return spec, Normalize(spec), nil
}

// Parse reads a JSON or YAML document, picking the format from the extension
// and trying both when it has none.
func Parse(path string) (*openapi31.Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read spec: %w", err)
	}
	var spec openapi31.Spec
	switch strings.ToLower(filepath.Ext(path)) {
	case ".json":
		if err := json.Unmarshal(data, &spec); err != nil {
			return nil, fmt.Errorf("parse OpenAPI JSON: %w", err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &spec); err != nil {
			return nil, fmt.Errorf("parse OpenAPI YAML: %w", err)
		}
	default:
		if jsonErr := json.Unmarshal(data, &spec); jsonErr != nil {
			if err := yaml.Unmarshal(data, &spec); err != nil {
				return nil, fmt.Errorf("parse OpenAPI file as JSON (%v) or YAML: %w", jsonErr, err)
			}
		}
	}
	return &spec, nil
}

// operation is one route of the document while it is being normalised.
type operation struct {
	Path   string
	Method string
	Op     *openapi31.Operation
}

// Normalize rewrites spec in place so that every operation left has a unique
// id and an `x-service`, and every `$ref` it carries resolves. It logs each
// decision it takes on a named operation, so a regeneration that silently
// changed the SDK's surface can be traced to the line that did it.
func Normalize(spec *openapi31.Spec) Report {
	var report Report
	if spec.Paths == nil {
		return report
	}
	schemas := map[string]bool{}
	if spec.Components != nil {
		for name := range spec.Components.Schemas {
			schemas[name] = true
		}
	}

	ops := collect(spec)
	report.Operations = len(ops)

	var kept []operation
	for _, op := range ops {
		if ref := danglingRef(op.Op, schemas); ref != "" {
			log.Printf("dropping %s %s (%s): references undefined schema %s", op.Method, op.Path, id(op.Op), ref)
			report.Dangling++
			drop(spec, op)
			continue
		}
		kept = append(kept, op)
	}

	kept = foldTwins(spec, kept, &report)
	assignIDs(kept, &report)
	assignServices(kept, &report)
	declarePathParameters(kept, &report)
	report.Kept = len(kept)
	return report
}

// collect lists the document's operations in a fixed order: paths sorted,
// methods in Methods order.
func collect(spec *openapi31.Spec) []operation {
	paths := make([]string, 0, len(spec.Paths.MapOfPathItemValues))
	for path := range spec.Paths.MapOfPathItemValues {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	var ops []operation
	for _, path := range paths {
		item := spec.Paths.MapOfPathItemValues[path]
		for _, method := range Methods {
			if op := byMethod(&item, method); op != nil {
				ops = append(ops, operation{Path: path, Method: method, Op: op})
			}
		}
	}
	return ops
}

func byMethod(item *openapi31.PathItem, method string) *openapi31.Operation {
	switch method {
	case "DELETE":
		return item.Delete
	case "GET":
		return item.Get
	case "PATCH":
		return item.Patch
	case "POST":
		return item.Post
	case "PUT":
		return item.Put
	}
	return nil
}

// drop removes an operation from the document, and the path item with it when
// no method is left.
func drop(spec *openapi31.Spec, op operation) {
	item := spec.Paths.MapOfPathItemValues[op.Path]
	switch op.Method {
	case "DELETE":
		item.Delete = nil
	case "GET":
		item.Get = nil
	case "PATCH":
		item.Patch = nil
	case "POST":
		item.Post = nil
	case "PUT":
		item.Put = nil
	}
	if item.Get == nil && item.Put == nil && item.Post == nil && item.Delete == nil && item.Patch == nil {
		delete(spec.Paths.MapOfPathItemValues, op.Path)
		return
	}
	spec.Paths.MapOfPathItemValues[op.Path] = item
}

func id(op *openapi31.Operation) string {
	if op.ID == nil {
		return ""
	}
	return *op.ID
}

func extension(op *openapi31.Operation, key string) string {
	if value, ok := op.MapOfAnything[key].(string); ok {
		return value
	}
	return ""
}

// danglingRef returns the first `#/components/schemas/…` reference of the
// operation that the document does not define, or "".
func danglingRef(op *openapi31.Operation, schemas map[string]bool) string {
	raw, err := json.Marshal(op)
	if err != nil {
		return ""
	}
	var generic any
	if err := json.Unmarshal(raw, &generic); err != nil {
		return ""
	}
	var refs []string
	walkRefs(generic, &refs)
	sort.Strings(refs)
	for _, ref := range refs {
		name, ok := strings.CutPrefix(ref, "#/components/schemas/")
		if !ok || !schemas[name] {
			return ref
		}
	}
	return ""
}

func walkRefs(value any, refs *[]string) {
	switch v := value.(type) {
	case map[string]any:
		for key, child := range v {
			if key == "$ref" {
				if ref, ok := child.(string); ok {
					*refs = append(*refs, ref)
				}
				continue
			}
			walkRefs(child, refs)
		}
	case []any:
		for _, child := range v {
			walkRefs(child, refs)
		}
	}
}

var parameterSegment = regexp.MustCompile(`\{[^}]+\}`)

// template is a path with its parameter names blanked: what two spellings of
// one route have in common.
func template(path string) string {
	return parameterSegment.ReplaceAllString(path, "{}")
}

// foldTwins keeps one operation per (method, template). Two paths that differ
// only by the name of a parameter are one route — the server does not see the
// name — and a document that carries both is a document that kept the old
// spelling when the parameter was renamed. The kept one is, in order, the one
// that declares an `x-service`, the one that has tags, the one whose path sorts
// first.
func foldTwins(spec *openapi31.Spec, ops []operation, report *Report) []operation {
	type key struct{ method, template string }
	groups := map[key][]operation{}
	var order []key
	for _, op := range ops {
		k := key{op.Method, template(op.Path)}
		if _, seen := groups[k]; !seen {
			order = append(order, k)
		}
		groups[k] = append(groups[k], op)
	}
	var kept []operation
	for _, k := range order {
		group := groups[k]
		if len(group) == 1 {
			kept = append(kept, group[0])
			continue
		}
		sort.SliceStable(group, func(i, j int) bool {
			a, b := group[i], group[j]
			if (extension(a.Op, "x-service") != "") != (extension(b.Op, "x-service") != "") {
				return extension(a.Op, "x-service") != ""
			}
			if (len(a.Op.Tags) > 0) != (len(b.Op.Tags) > 0) {
				return len(a.Op.Tags) > 0
			}
			return a.Path < b.Path
		})
		kept = append(kept, group[0])
		for _, twin := range group[1:] {
			log.Printf("dropping %s %s (%s): same route as %s (%s)", twin.Method, twin.Path, id(twin.Op), group[0].Path, id(group[0].Op))
			report.Twins++
			drop(spec, twin)
		}
	}
	return kept
}

// assignIDs gives every operation a unique id: the override table first, then
// the document's own, then — for an operation that has none or whose id another
// operation already uses — one derived from the path. A derived id is logged
// with the override key that would replace it.
func assignIDs(ops []operation, report *Report) {
	for _, op := range ops {
		if override, ok := naming.OperationIDOverrides[op.Method+":"+op.Path]; ok {
			op.Op.ID = &override
		}
	}
	byID := map[string][]operation{}
	for _, op := range ops {
		byID[id(op.Op)] = append(byID[id(op.Op)], op)
	}
	taken := map[string]bool{}
	for name := range byID {
		if name != "" {
			taken[name] = true
		}
	}
	names := make([]string, 0, len(byID))
	for name := range byID {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		group := byID[name]
		if name != "" && len(group) == 1 {
			continue
		}
		// The first occurrence keeps the id; the others get one derived from
		// what tells them apart. An operation with no id at all derives one from
		// its whole route.
		for i, op := range group {
			if name != "" && i == 0 {
				continue
			}
			derived := deriveID(op, name, group)
			for taken[derived] {
				derived += naming.PascalCase(op.Method)
			}
			taken[derived] = true
			op.Op.ID = &derived
			report.Renamed++
			log.Printf("renaming %s %s: operationId %q is taken, using %q (add %q to naming.OperationIDOverrides to choose)",
				op.Method, op.Path, name, derived, op.Method+":"+op.Path)
		}
	}
}

// deriveID is the mechanical id of an operation whose own is missing or taken:
// the taken id followed by the first path segment that differs from the
// operation that keeps the id (`createInternalApplication` on `/internal/…`
// when the id belongs to `/v2/internal/…`), or, with no id to start from, the
// method and the path's literal segments.
func deriveID(op operation, name string, group []operation) string {
	if name == "" {
		return naming.PascalCase(strings.ToLower(op.Method) + "_" + strings.ReplaceAll(strings.ReplaceAll(template(op.Path), "{}", ""), "/", "_"))
	}
	mine := strings.Split(strings.Trim(op.Path, "/"), "/")
	theirs := strings.Split(strings.Trim(group[0].Path, "/"), "/")
	for i, segment := range mine {
		if strings.HasPrefix(segment, "{") {
			continue
		}
		if i >= len(theirs) || theirs[i] != segment {
			return name + naming.PascalCase(segment)
		}
	}
	return name + naming.PascalCase(op.Method)
}

// assignServices fills `x-service` on every operation that has none: from its
// first tag, or from its path.
func assignServices(ops []operation, report *Report) {
	for _, op := range ops {
		if extension(op.Op, "x-service") != "" {
			continue
		}
		service := ""
		if len(op.Op.Tags) > 0 {
			service = op.Op.Tags[0]
		} else {
			service = serviceFromPath(op.Path)
		}
		if op.Op.MapOfAnything == nil {
			op.Op.MapOfAnything = map[string]interface{}{}
		}
		op.Op.MapOfAnything["x-service"] = service
		report.Tagged++
	}
}

// declarePathParameters adds to each operation the path parameters its template
// names but its `parameters` do not declare, as required strings. A generator
// formats every `{segment}` of the path from an argument; a route that names
// one it never declared would get a function with one argument too few and a
// builder calling it with one too many, and the SDK would not compile. The
// document carried 45 such loadbalancer routes once (fixed upstream) and
// carries two dozen es-addon ones today.
//
// A declared parameter counts for a template segment of the same name, of the
// same name up to case and underscores (`addon_id` for `{addonId}`), or, once
// those are paired off, by position: a template with two segments and two
// declared path parameters is fully declared whatever they are called.
func declarePathParameters(ops []operation, report *Report) {
	for _, op := range ops {
		var declared []string
		for _, p := range op.Op.Parameters {
			if p.Parameter != nil && p.Parameter.In == openapi31.ParameterInPath {
				declared = append(declared, p.Parameter.Name)
			}
		}
		var segments []string
		for _, match := range parameterSegment.FindAllString(op.Path, -1) {
			segments = append(segments, strings.Trim(match, "{}"))
		}
		unmatched := unmatchedSegments(segments, declared)
		for i, name := range unmatched {
			if i < len(declared)-(len(segments)-len(unmatched)) {
				// Paired by position with a declared parameter of another name.
				continue
			}
			required := true
			op.Op.Parameters = append(op.Op.Parameters, openapi31.ParameterOrReference{Parameter: &openapi31.Parameter{
				Name:     name,
				In:       openapi31.ParameterInPath,
				Required: &required,
				Schema:   map[string]interface{}{"type": "string"},
			}})
			report.Declared++
			log.Printf("declaring path parameter %q on %s %s (%s): the template names it, the operation does not", name, op.Method, op.Path, id(op.Op))
		}
	}
}

// unmatchedSegments returns the template segments no declared parameter names,
// exactly or up to case and underscores.
func unmatchedSegments(segments, declared []string) []string {
	fold := func(s string) string {
		return strings.ToLower(strings.NewReplacer("_", "", "-", "").Replace(s))
	}
	used := map[int]bool{}
	var unmatched []string
	for _, segment := range segments {
		found := false
		for i, name := range declared {
			if !used[i] && (name == segment || fold(name) == fold(segment)) {
				used[i] = true
				found = true
				break
			}
		}
		if !found {
			unmatched = append(unmatched, segment)
		}
	}
	return unmatched
}

// serviceFromPath names a service after a path: the provider of an
// `/addon-providers/addon-foo/…` route, otherwise the first segment after the
// version prefix.
func serviceFromPath(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for i, part := range parts {
		if part == "addon-providers" && i+1 < len(parts) {
			return strings.TrimPrefix(parts[i+1], "addon-")
		}
	}
	for _, part := range parts {
		if part == "v2" || part == "v4" || strings.HasPrefix(part, "{") {
			continue
		}
		return part
	}
	return "unknown"
}
