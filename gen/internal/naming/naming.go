// Package naming derives the Go identifiers the SDK publishes from the names the
// OpenAPI document carries: service packages from `x-service` values and tags,
// function names from operation ids.
//
// It is the single source of these rules. The three generators used to carry
// their own copies of the acronym table, the package exceptions and the
// PascalCase routine, and every regeneration found a new place where the copies
// had drifted: a builder importing `services/NetworkGroup` while the directory
// on disk was `services/network_group`, a builder calling
// `GetInternalBestAppForUri` where the package exported `…ForURI`. A rule
// implemented twice with nothing keeping the copies honest is exactly the
// defect this package removes.
package naming

import (
	"sort"
	"strings"
)

// ServicePackageExceptions maps an `x-service` value or a tag to the package it
// is generated into when the mechanical derivation (see PackageName) would not
// do. Keys are matched exactly, then case-insensitively.
var ServicePackageExceptions = map[string]string{
	// `internal` is reserved by Go: a package at services/internal is importable only from
	// services/..., so builder.go at the module root cannot reach it and the SDK stops
	// compiling. AXO publishes an `Internal` tag, hence the rename.
	"internal": "internalapi",

	// `vendor` is reserved by the go command: directories of that name are stripped from a
	// module zip, so `go get` of this SDK yields a module with no services/vendor package at
	// all and every consumer fails to build. AXO publishes a `Vendor` tag, hence the rename.
	"vendor": "vendorapi",

	// Addon providers published under their legacy `addon-*` tags.
	"addon-pulsar":     "pulsar",
	"addon-storage":    "storage",
	"addon-cellar":     "storage",
	"addon-ai":         "ai",
	"addon-cumulocity": "cumulocity",
	"addon-keycloak":   "keycloak",
	"addon-matomo":     "matomo",
	"addon-metabase":   "metabase",
	"addon-otoroshi":   "otoroshi",
	"addon-postgresql": "postgresql",
	"config-provider":  "configuration_provider",
	"network-group":    "network_group",
	"materia-kv":       "materia_kv",
	"compute":          "infrastructure",
	"drain":            "drains",

	// Tags the unified document carries on operations that have no `x-service`. The
	// mechanical derivation splits CamelCase at every capital, which spells `PostgreSQL`
	// as `postgre_s_q_l`; product names with embedded acronyms are listed instead.
	"postgresql internal":    "postgresql_internal",
	"postgresql dashboard":   "postgresql_dashboard",
	"mysql internal":         "mysql_internal",
	"mysql dashboard":        "mysql_dashboard",
	"mysql addon":            "my_sql_addon",
	"mongodb internal":       "mongodb_internal",
	"mongodb dashboard":      "mongodb_dashboard",
	"mongodb addon":          "mongodb_addon",
	"elasticsearch internal": "elasticsearch_internal",
	"elasticsearch addon":    "elasticsearch_addon",
	"es-addon":               "elasticsearch_addon",
	"jenkins addon":          "jenkins_addon",
	"hivemq":                 "hivemq",
	"drains v4":              "drains",
	"warp10token":            "warp10_token",
	"warp10 token":           "warp10_token",
	"functions-krun":         "functions_krun",
	// Tags that only differ from an `x-service` value by their casing, so the two name one
	// package rather than `loadbalancer` and `load_balancer` side by side.
	"loadbalancer":     "loadbalancer",
	"loadbalancerinfo": "loadbalancer_info",
	"postgresql addon": "postgresql_addon",
	"token":            "tokens",
}

// OperationIDOverrides renames operations the document publishes under an id
// another operation already uses. Keys are `METHOD:/path`, exactly as the
// document spells the path.
//
// Collisions this table does not cover are resolved mechanically by
// spec.Normalize, which logs each one; add an entry here when the mechanical
// name is not the one the SDK should publish.
var OperationIDOverrides = map[string]string{
	// Fix duplicate "createEndpoint" operationId
	"DELETE:/v4/ai/organisations/{ownerId}/ai/{addonAIId}/endpoints/{endpointId}": "deleteAIEndpoint",
	"POST:/v4/ai/organisations/{ownerId}/ai/{addonAIId}/endpoints":                "createAIEndpoint",
	// Fix duplicate getMateriaKv/deleteMateriaKv (exist on both /v2 and /v4 paths)
	"GET:/v4/materia/organisations/{ownerId}/materia/databases/{resourceId}":    "getMateriaKvV4",
	"DELETE:/v4/materia/organisations/{ownerId}/materia/databases/{resourceId}": "deleteMateriaKvV4",
}

// OperationPackageExceptions moves individual operations to a package other than
// the one their `x-service` names. See SERVICE_MAPPING_EXCEPTIONS.md.
var OperationPackageExceptions = map[string]string{
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

// acronyms Go spells in full capitals. Applied by FixIDSuffixes at the end of a
// word or before another capital only: a blanket replacement would rewrite `Ids`
// into `IDs` and `Idle` into `IDle`.
var acronyms = map[string]string{
	"Api":   "API",
	"Http":  "HTTP",
	"Https": "HTTPS",
	"Id":    "ID",
	"Ip":    "IP",
	"Ipam":  "IPAM",
	"Json":  "JSON",
	"Sql":   "SQL",
	"Tls":   "TLS",
	"Ttl":   "TTL",
	"Uri":   "URI",
	"Url":   "URL",
	"Uuid":  "UUID",
	"Xml":   "XML",
	"Yaml":  "YAML",
}

// sortedAcronyms is the acronym table in a fixed order: map iteration is
// randomised per run, and a generator must produce the same output every time.
var sortedAcronyms = func() []string {
	keys := make([]string, 0, len(acronyms))
	for k := range acronyms {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}()

// FixIDSuffixes upper-cases the acronyms of an identifier: `GetId` is `GetID`,
// `AppIdFor` is `AppIDFor`, `GetIds` stays `GetIds`.
func FixIDSuffixes(s string) string {
	for _, old := range sortedAcronyms {
		new := acronyms[old]
		if strings.HasSuffix(s, old) {
			s = strings.TrimSuffix(s, old) + new
			continue
		}
		for i := 'A'; i <= 'Z'; i++ {
			next := string(i)
			s = strings.ReplaceAll(s, old+next, new+next)
		}
	}
	return s
}

// PascalCase turns an operation id, a path segment or a parameter name into an
// exported identifier. Words are split on `_`, `-`, `.` and spaces; braces are
// dropped wherever they sit, because a path segment can mix a parameter with a
// literal suffix (`{bid}.pdf` is a real route) and `strings.Trim` only strips the
// ends. Each word is capitalised and the rest lower-cased, then the acronym
// table is applied.
func PascalCase(s string) string {
	s = strings.ReplaceAll(s, "{", "")
	s = strings.ReplaceAll(s, "}", "")
	words := strings.FieldsFunc(s, func(c rune) bool {
		return c == '_' || c == '-' || c == ' ' || c == '.'
	})
	var result strings.Builder
	for _, word := range words {
		result.WriteString(strings.ToUpper(word[:1]))
		result.WriteString(strings.ToLower(word[1:]))
	}
	return FixIDSuffixes(result.String())
}

// FunctionName is the exported function an operation id becomes, in the
// services package and in the builder alike.
func FunctionName(operationID string) string {
	return PascalCase(operationID)
}

// SnakeCase lower-cases a CamelCase name, splitting words at a capital that
// follows a lower-case letter or a digit, or that starts a word inside a run of
// capitals; hyphens and spaces become underscores so the result is a legal
// package name. `NetworkGroup` is `network_group`, `Warp10Token` is
// `warp10_token`, `IPAM` is `ipam`, `functions-krun` is `functions_krun`.
func SnakeCase(s string) string {
	runes := []rune(s)
	var result strings.Builder
	for i, r := range runes {
		switch {
		case r >= 'A' && r <= 'Z':
			if i > 0 {
				prev := runes[i-1]
				prevLowerOrDigit := (prev >= 'a' && prev <= 'z') || (prev >= '0' && prev <= '9')
				prevUpper := prev >= 'A' && prev <= 'Z'
				nextLower := i+1 < len(runes) && runes[i+1] >= 'a' && runes[i+1] <= 'z'
				if prevLowerOrDigit || (prevUpper && nextLower) {
					result.WriteRune('_')
				}
			}
			result.WriteRune(r + ('a' - 'A'))
		case r == '-' || r == ' ' || r == '.':
			result.WriteRune('_')
		default:
			result.WriteRune(r)
		}
	}
	return strings.Trim(strings.ReplaceAll(result.String(), "__", "_"), "_")
}

// PackageName is the services package an `x-service` value or tag is generated
// into: the exception table first, then SnakeCase.
func PackageName(service string) string {
	if pkg, ok := ServicePackageExceptions[service]; ok {
		return pkg
	}
	if pkg, ok := ServicePackageExceptions[strings.ToLower(service)]; ok {
		return pkg
	}
	return SnakeCase(service)
}

// PulsarPackage applies the rules of SERVICE_MAPPING_EXCEPTIONS.md: every
// operation that mentions pulsar in its path, its id or its tags belongs to the
// pulsar package, whatever its `x-service` says. It returns "" when none applies.
func PulsarPackage(operationID, path string, tags []string) string {
	if strings.Contains(strings.ToLower(path), "pulsar") {
		return "pulsar"
	}
	if strings.Contains(strings.ToLower(operationID), "pulsar") {
		return "pulsar"
	}
	for _, tag := range tags {
		if strings.ToLower(tag) == "pulsar" {
			return "pulsar"
		}
	}
	return ""
}

// OperationPackage is the package an operation is generated into: the
// per-operation exceptions, then the pulsar rules, then its `x-service` (or, when
// it has none, its first tag) through PackageName. Both generators call this and
// nothing else, so they cannot disagree.
func OperationPackage(operationID, xService, path string, tags []string) string {
	if pkg, ok := OperationPackageExceptions[operationID]; ok {
		return pkg
	}
	if pkg := PulsarPackage(operationID, path, tags); pkg != "" {
		return pkg
	}
	if xService == "" && len(tags) > 0 {
		xService = tags[0]
	}
	if xService == "" {
		return ""
	}
	return PackageName(xService)
}
