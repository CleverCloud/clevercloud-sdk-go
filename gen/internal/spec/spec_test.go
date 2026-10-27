package spec

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/swaggest/openapi-go/openapi31"
)

// load parses a document given as JSON and normalises it.
func load(t *testing.T, document string) (*openapi31.Spec, Report) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "spec.json")
	document = `{"openapi":"3.1.0","info":{"title":"test","version":"1"},` + strings.TrimPrefix(strings.TrimSpace(document), "{")
	if err := os.WriteFile(path, []byte(document), 0o644); err != nil {
		t.Fatal(err)
	}
	spec, report, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	return spec, report
}

func operations(spec *openapi31.Spec) map[string]string {
	ids := map[string]string{}
	for _, op := range collect(spec) {
		ids[op.Method+" "+op.Path] = id(op.Op)
	}
	return ids
}

func TestDanglingReferencesAreDropped(t *testing.T) {
	spec, report := load(t, `{"paths":{
		"/v4/things": {
			"get": {"operationId": "listThings", "x-service": "things", "responses": {"200": {"description": "ok", "content": {"application/json": {"schema": {"type": "array", "items": {"$ref": "#/components/schemas/Thing"}}}}}}},
			"post": {"operationId": "createThing", "x-service": "things", "requestBody": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/WannabeThing"}}}}}
		}
	}, "components": {"schemas": {"Thing": {"type": "object"}}}}`)
	if report.Dangling != 1 || report.Kept != 1 {
		t.Fatalf("report = %+v", report)
	}
	if spec.Paths.MapOfPathItemValues["/v4/things"].Post != nil {
		t.Error("the operation with a dangling $ref survived")
	}
	if spec.Paths.MapOfPathItemValues["/v4/things"].Get == nil {
		t.Error("the operation whose $ref resolves was dropped")
	}
}

func TestTwinSpellingsAreFolded(t *testing.T) {
	spec, report := load(t, `{"paths":{
		"/v2/organisations/{id}/addons/{addon_id}": {"get": {"operationId": "get_addon"}},
		"/v2/organisations/{id}/addons/{addonId}": {"get": {"operationId": "getAddon", "x-service": "addons"}, "delete": {"operationId": "deleteAddon", "x-service": "addons"}}
	}}`)
	if report.Twins != 1 || report.Kept != 2 {
		t.Fatalf("report = %+v", report)
	}
	got := operations(spec)
	if got["GET /v2/organisations/{id}/addons/{addonId}"] != "getAddon" {
		t.Errorf("the spelling with an x-service must be kept, got %v", got)
	}
	if _, kept := got["GET /v2/organisations/{id}/addons/{addon_id}"]; kept {
		t.Errorf("the other spelling survived: %v", got)
	}
}

func TestDuplicateIDsAreRenamed(t *testing.T) {
	spec, report := load(t, `{"paths":{
		"/internal/invoices": {"get": {"operationId": "get_internal_invoices", "tags": ["Internal"]}},
		"/v2/internal/invoices": {"get": {"operationId": "get_internal_invoices", "x-service": "internal"}},
		"/v4/materia/organisations/{ownerId}/materia/databases/{resourceId}": {"get": {"operationId": "getMateriaKv", "x-service": "materia_kv"}},
		"/v2/providers/kv/resources/{id}": {"get": {"operationId": "getMateriaKv", "x-service": "materia_kv"}},
		"/v4/noid": {"get": {"x-service": "misc"}}
	}}`)
	if report.Renamed != 2 {
		t.Fatalf("report = %+v", report)
	}
	got := operations(spec)
	want := map[string]string{
		"GET /internal/invoices":    "get_internal_invoices",
		"GET /v2/internal/invoices": "get_internal_invoicesV2",
		"GET /v4/materia/organisations/{ownerId}/materia/databases/{resourceId}": "getMateriaKvV4",
		"GET /v2/providers/kv/resources/{id}":                                    "getMateriaKv",
		"GET /v4/noid":                                                           "GetV4Noid",
	}
	for route, id := range want {
		if got[route] != id {
			t.Errorf("%s: got id %q, want %q", route, got[route], id)
		}
	}
	ids := map[string]bool{}
	for _, id := range got {
		if ids[id] {
			t.Errorf("id %q published twice", id)
		}
		ids[id] = true
	}
}

func TestServicesAreDerived(t *testing.T) {
	spec, report := load(t, `{"paths":{
		"/v4/billing/invoices": {"get": {"operationId": "a", "tags": ["Billing"]}},
		"/v4/addon-providers/addon-foo/addons": {"get": {"operationId": "b"}},
		"/v2/self/tokens": {"get": {"operationId": "c"}},
		"/v4/kept": {"get": {"operationId": "d", "x-service": "kept", "tags": ["Other"]}}
	}}`)
	if report.Tagged != 3 {
		t.Fatalf("report = %+v", report)
	}
	want := map[string]string{
		"/v4/billing/invoices":                 "Billing",
		"/v4/addon-providers/addon-foo/addons": "foo",
		"/v2/self/tokens":                      "self",
		"/v4/kept":                             "kept",
	}
	for path, service := range want {
		if got := extension(spec.Paths.MapOfPathItemValues[path].Get, "x-service"); got != service {
			t.Errorf("%s: x-service %q, want %q", path, got, service)
		}
	}
}

func TestMissingPathParametersAreDeclared(t *testing.T) {
	spec, report := load(t, `{"paths":{
		"/v2/es/{id}/configuration": {"get": {"operationId": "a", "x-service": "es"}},
		"/v2/organisations/{id}/addons/{addonId}/sso": {"post": {"operationId": "b", "x-service": "addons", "parameters": [
			{"name": "id", "in": "path", "required": true, "schema": {"type": "string"}},
			{"name": "addon_id", "in": "path", "required": true, "schema": {"type": "string"}}]}},
		"/v4/pg/{ownerId}/users/{pgUserId}": {"get": {"operationId": "c", "x-service": "pg", "parameters": [
			{"name": "owner_id", "in": "path", "required": true, "schema": {"type": "string"}},
			{"name": "user_id", "in": "path", "required": true, "schema": {"type": "string"}}]}}
	}}`)
	if report.Declared != 1 {
		t.Fatalf("report = %+v", report)
	}
	params := spec.Paths.MapOfPathItemValues["/v2/es/{id}/configuration"].Get.Parameters
	if len(params) != 1 || params[0].Parameter.Name != "id" || params[0].Parameter.In != openapi31.ParameterInPath {
		raw, _ := json.Marshal(params)
		t.Errorf("declared parameters = %s", raw)
	}
	for _, path := range []string{"/v2/organisations/{id}/addons/{addonId}/sso", "/v4/pg/{ownerId}/users/{pgUserId}"} {
		item := spec.Paths.MapOfPathItemValues[path]
		op := item.Post
		if op == nil {
			op = item.Get
		}
		if len(op.Parameters) != 2 {
			t.Errorf("%s: %d parameters, want the 2 it declared under other names", path, len(op.Parameters))
		}
	}
}
