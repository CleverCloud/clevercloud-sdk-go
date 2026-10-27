package gotype

import (
	"bytes"
	"strings"
	"testing"

	"github.com/dave/jennifer/jen"
	"github.com/swaggest/openapi-go/openapi31"
)

func content(mediaType string, schema map[string]any) map[string]openapi31.MediaType {
	return map[string]openapi31.MediaType{mediaType: {Schema: schema}}
}

func ref(name string) map[string]any {
	return map[string]any{"$ref": "#/components/schemas/" + name}
}

func TestBodyShapes(t *testing.T) {
	cases := []struct {
		name   string
		schema map[string]any
		want   string
	}{
		{"component", ref("WannabeAddon"), "*models.WannabeAddon"},
		{"nullable component", map[string]any{"oneOf": []any{map[string]any{"type": "null"}, ref("SecondaryEmailPayload")}}, "*models.SecondaryEmailPayload"},
		{"array of components", map[string]any{"type": "array", "items": ref("Tag")}, "[]*models.Tag"},
		{"map of strings", map[string]any{"type": "object", "additionalProperties": map[string]any{"type": "string"}}, "map[string]string"},
		{"array of strings", map[string]any{"type": "array", "items": map[string]any{"type": "string"}}, "[]string"},
		{"empty schema", map[string]any{}, "any"},
		{"free object", map[string]any{"type": "object"}, "map[string]any"},
		{"string", map[string]any{"type": "string"}, "string"},
	}
	for _, c := range cases {
		if got := Body(content("application/json", c.schema), "Op"); got.Expr != c.want {
			t.Errorf("%s: Body = %q, want %q", c.name, got.Expr, c.want)
		}
	}
}

func TestResponseShapes(t *testing.T) {
	cases := []struct {
		name      string
		mediaType string
		schema    map[string]any
		want      string
	}{
		{"component by value", "application/json", ref("AddonView"), "models.AddonView"},
		{"array by value", "application/json", map[string]any{"type": "array", "items": ref("AddonView")}, "[]models.AddonView"},
		{"map of arrays", "application/json", map[string]any{"type": "object", "additionalProperties": map[string]any{"type": "array", "items": ref("DeploymentView")}}, "map[string][]models.DeploymentView"},
		{"nullable string", "application/json", map[string]any{"type": []any{"string", "null"}}, "*string"},
		{"text is not decodable", "text/plain", map[string]any{"type": "string"}, Nothing.Expr},
		{"html is not decodable", "text/html", nil, Nothing.Expr},
	}
	for _, c := range cases {
		if got := Response(content(c.mediaType, c.schema), "Op"); got.Expr != c.want {
			t.Errorf("%s: Response = %q, want %q", c.name, got.Expr, c.want)
		}
	}
	if got := Response(nil, "Op"); !got.IsNothing() {
		t.Errorf("no content: Response = %q, want Nothing", got.Expr)
	}
}

func TestInlineObjectBecomesANamedStruct(t *testing.T) {
	schema := map[string]any{
		"type":        "object",
		"description": "New name of the add-on.",
		"required":    []any{"name"},
		"properties": map[string]any{
			"name":  map[string]any{"type": "string", "description": "Not blank."},
			"tags":  map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"owner": map[string]any{"type": "object", "properties": map[string]any{"id": map[string]any{"type": "string"}}},
		},
	}
	got := Body(content("application/json", schema), "Updateaddoninfo")
	if got.Expr != "UpdateaddoninfoBody" {
		t.Fatalf("Expr = %q", got.Expr)
	}
	if len(got.Inline) != 2 || got.Inline[0].Name != "UpdateaddoninfoBodyOwner" || got.Inline[1].Name != "UpdateaddoninfoBody" {
		t.Fatalf("Inline = %+v", got.Inline)
	}
	body := got.Inline[1]
	want := map[string]string{"Name": "string", "Owner": "*UpdateaddoninfoBodyOwner", "Tags": "[]string"}
	for _, f := range body.Fields {
		if want[f.Name] != f.Expr {
			t.Errorf("field %s: %q, want %q", f.Name, f.Expr, want[f.Name])
		}
	}
	var buf bytes.Buffer
	if err := jen.NewFile("x").Add(body.Declare()).Render(&buf); err != nil {
		t.Fatal(err)
	}
	rendered := strings.Join(strings.Fields(buf.String()), " ")
	for _, snippet := range []string{"type UpdateaddoninfoBody struct", "Name string `json:\"name\"`", "Tags []string `json:\"tags,omitempty\"`", "Owner *UpdateaddoninfoBodyOwner `json:\"owner,omitempty\"`"} {
		if !strings.Contains(rendered, snippet) {
			t.Errorf("declaration lacks %q:\n%s", snippet, buf.String())
		}
	}
}

func TestPickPrefersJSONDeterministically(t *testing.T) {
	c := map[string]openapi31.MediaType{
		"text/plain":                  {Schema: map[string]any{"type": "string"}},
		"application/json":            {Schema: ref("A")},
		"application/json-patch+json": {Schema: ref("B")},
	}
	for i := 0; i < 20; i++ {
		if got := Response(c, "Op"); got.Expr != "models.A" {
			t.Fatalf("Response = %q, want models.A", got.Expr)
		}
	}
	delete(c, "application/json")
	if got := Response(c, "Op"); got.Expr != "models.B" {
		t.Fatalf("Response = %q, want models.B", got.Expr)
	}
}

func TestCodeRendersQualifiedTypes(t *testing.T) {
	inline := func(name string) jen.Code { return jen.Qual("go.clever-cloud.dev/sdk/services/addons", name) }
	cases := map[string]string{
		"*models.WannabeAddon":  "*models.WannabeAddon",
		"[]models.AddonView":    "[]models.AddonView",
		"map[string]string":     "map[string]string",
		"UpdateaddoninfoBody":   "addons.UpdateaddoninfoBody",
		"client.Nothing":        "client.Nothing",
		"any":                   "any",
		"map[string][]models.X": "map[string][]models.X",
	}
	for expr, want := range cases {
		var buf bytes.Buffer
		if err := jen.NewFile("x").Var().Id("v").Add(Type{Expr: expr}.Code(inline)).Render(&buf); err != nil {
			t.Fatal(err)
		}
		if !bytes.Contains(buf.Bytes(), []byte("var v "+want)) {
			t.Errorf("%s rendered as\n%s", expr, buf.String())
		}
	}
}
