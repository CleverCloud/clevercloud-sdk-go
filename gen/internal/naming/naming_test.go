package naming

import "testing"

func TestPascalCase(t *testing.T) {
	cases := map[string]string{
		"listKubernetesClusters":         "Listkubernetesclusters",
		"create_internal_application":    "CreateInternalApplication",
		"get_internal_best_app_for_uri":  "GetInternalBestAppForURI",
		"postgresql_internal_ids_by_own": "PostgresqlInternalIdsByOwn",
		"addon-providers":                "AddonProviders",
		"{bid}.pdf":                      "BidPdf",
		"getAppId":                       "Getappid",
		"get_app_id":                     "GetAppID",
		"get_app_ids":                    "GetAppIds",
		"list_http_endpoints":            "ListHTTPEndpoints",
	}
	for in, want := range cases {
		if got := PascalCase(in); got != want {
			t.Errorf("PascalCase(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSnakeCase(t *testing.T) {
	cases := map[string]string{
		"NetworkGroup":     "network_group",
		"Warp10Token":      "warp10_token",
		"IPAM":             "ipam",
		"functions-krun":   "functions_krun",
		"OpenTelemetry":    "open_telemetry",
		"LoadBalancerInfo": "load_balancer_info",
		"MySQL":            "my_sql",
		"Billing Internal": "billing_internal",
		"cellar":           "cellar",
	}
	for in, want := range cases {
		if got := SnakeCase(in); got != want {
			t.Errorf("SnakeCase(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestPackageName(t *testing.T) {
	cases := map[string]string{
		"Internal":               "internalapi",
		"internal":               "internalapi",
		"Vendor":                 "vendorapi",
		"PostgreSQL Internal":    "postgresql_internal",
		"MySQL Addon":            "my_sql_addon",
		"es-addon":               "elasticsearch_addon",
		"Drains v4":              "drains",
		"postgresql_addon":       "postgresql_addon",
		"ContainerRegistry":      "container_registry",
		"configuration_provider": "configuration_provider",
	}
	for in, want := range cases {
		if got := PackageName(in); got != want {
			t.Errorf("PackageName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestOperationPackage(t *testing.T) {
	cases := []struct {
		id, service, path string
		tags              []string
		want              string
	}{
		{"getPulsar", "products", "/v4/addon-providers/addon-pulsar/addons/{id}", nil, "pulsar"},
		{"listThings", "misc", "/v4/pulsar/things", nil, "pulsar"},
		{"listThings", "misc", "/v4/things", []string{"Pulsar"}, "pulsar"},
		{"listThings", "network_group", "/v4/things", nil, "network_group"},
		{"listThings", "", "/v4/things", []string{"Billing Internal"}, "billing_internal"},
		{"listThings", "", "/v4/things", nil, ""},
	}
	for _, c := range cases {
		if got := OperationPackage(c.id, c.service, c.path, c.tags); got != c.want {
			t.Errorf("OperationPackage(%q, %q, %q, %v) = %q, want %q", c.id, c.service, c.path, c.tags, got, c.want)
		}
	}
}
