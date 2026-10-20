package app

import (
	"slices"
	"strings"
	"testing"
)

func TestLookupString(t *testing.T) {
	t.Setenv("TEST_STRING", "  value  ")
	if got := lookupString("def", false, "TEST_STRING"); got != "value" {
		t.Errorf("got %q, want value", got)
	}

	t.Setenv("TEST_EMPTY", " ")
	if got := lookupString("def", false, "TEST_EMPTY"); got != "def" {
		t.Errorf("empty value: got %q, want the default", got)
	}
	if got := lookupString("", true, "TEST_EMPTY"); got != "" {
		t.Errorf("empty value allowed: got %q, want empty", got)
	}
	if got := lookupString("def", false, "TEST_UNSET"); got != "def" {
		t.Errorf("unset: got %q, want the default", got)
	}
}

func TestLookupBool(t *testing.T) {
	for value, want := range map[string]bool{"1": true, "Enabled": true, "install": true, "OFF": false, "skip": false} {
		t.Setenv("TEST_BOOL", value)
		got, err := lookupBool(false, []string{"install"}, []string{"skip"}, "TEST_BOOL")
		if err != nil || got != want {
			t.Errorf("%q: got %v (%v), want %v", value, got, err, want)
		}
	}

	t.Setenv("TEST_BOOL", "maybe")
	if _, err := lookupBool(false, nil, nil, "TEST_BOOL"); err == nil {
		t.Error("maybe: want an error")
	}
	if got, err := lookupBool(true, nil, nil, "TEST_UNSET"); err != nil || !got {
		t.Errorf("unset: got %v (%v), want the default true", got, err)
	}
}

func TestLookupDisabler(t *testing.T) {
	for value, want := range map[string]bool{"true": true, "off": true, "disable": true, "on": false, "enable": false, "0": false} {
		t.Setenv("TEST_DISABLER", value)
		got, err := lookupDisabler(false, "TEST_DISABLER")
		if err != nil || got != want {
			t.Errorf("%q: got %v (%v), want %v", value, got, err, want)
		}
	}
}

func TestLookupInt(t *testing.T) {
	t.Setenv("TEST_INT", " 42 ")
	if got, err := lookupInt(0, "TEST_INT"); err != nil || got != 42 {
		t.Errorf("got %d (%v), want 42", got, err)
	}
	for _, value := range []string{"007", "+1", "4 2", "1.5", "abc"} {
		t.Setenv("TEST_INT", value)
		if _, err := lookupInt(0, "TEST_INT"); err == nil {
			t.Errorf("%q: want an error", value)
		}
	}
	if got, err := lookupInt(300, "TEST_UNSET"); err != nil || got != 300 {
		t.Errorf("unset: got %d (%v), want the default 300", got, err)
	}
}

func TestLookupChoice(t *testing.T) {
	values := []string{"8", "21", "graalvm-ce"}
	t.Setenv("TEST_CHOICE", " GraalVM-CE ")
	if got, err := lookupChoice("21", values, "TEST_CHOICE"); err != nil || got != "graalvm-ce" {
		t.Errorf("got %q (%v), want graalvm-ce", got, err)
	}
	t.Setenv("TEST_CHOICE", "9")
	if _, err := lookupChoice("21", values, "TEST_CHOICE"); err == nil {
		t.Error("9: want an error")
	}
	if got, err := lookupChoice("21", values, "TEST_UNSET"); err != nil || got != "21" {
		t.Errorf("unset: got %q (%v), want the default 21", got, err)
	}
}

func TestLookupMultiple(t *testing.T) {
	t.Setenv("TEST_PORT", "1194")
	t.Setenv("TEST_PORT_10", "1300")
	t.Setenv("TEST_PORT_2", "1200")
	t.Setenv("TEST_PORT_1", "1100")
	t.Setenv("TEST_PORTS", "0")          // another variable, not an occurrence
	t.Setenv("TEST_PORT_X", "0")         // not an index
	t.Setenv("TEST_PORT_2_EXTRA", "nan") // not an index either

	entries, err := lookupMultiple(parseInt, "TEST_PORT")
	if err != nil {
		t.Fatal(err)
	}
	var got []int
	for _, e := range entries {
		got = append(got, e.Value)
	}
	if want := []int{1194, 1100, 1200, 1300}; !slices.Equal(got, want) {
		t.Errorf("got %v, want %v (sorted by numeric index)", got, want)
	}
	if entries[3].Key != "TEST_PORT_10" || entries[3].Index != 10 {
		t.Errorf("last entry = %+v, want TEST_PORT_10 at index 10", entries[3])
	}

	t.Setenv("TEST_PORT_3", "x")
	if _, err := lookupMultiple(parseInt, "TEST_PORT"); err == nil {
		t.Error("invalid occurrence: want an error")
	}
}

// The generated getters wire the right parser and default.
func TestGeneratedGetters(t *testing.T) {
	if got := MustGetJavaVersion(); got != "21" {
		t.Errorf("MustGetJavaVersion unset: got %q, want the default 21", got)
	}
	t.Setenv(CC_JAVA_VERSION, "17")
	if got := MustGetJavaVersion(); got != "17" {
		t.Errorf("MustGetJavaVersion: got %q, want 17", got)
	}
	t.Setenv("JAVA_VERSION", "11") // former name, ignored
	if got := MustGetJavaVersion(); got != "17" {
		t.Errorf("MustGetJavaVersion with the former name set: got %q, want 17", got)
	}

	t.Setenv(CC_NODE_DEV_DEPENDENCIES, "install")
	if got := MustGetNodeDevDependencies(); !got {
		t.Errorf("MustGetNodeDevDependencies(install): got %v, want true", got)
	}

	t.Setenv(CC_RUST_FEATURES, "tls,metrics")
	if got := GetRustFeatures(); got != "tls,metrics" {
		t.Errorf("GetRustFeatures: got %q", got)
	}

	t.Setenv(CC_WORKER_COMMAND, "./worker")
	if got := GetWorkerCommand(); len(got) != 1 || got[0].Value != "./worker" {
		t.Errorf("GetWorkerCommand: got %+v", got)
	}
}

func TestMustGetPanicsOnInvalidValue(t *testing.T) {
	t.Setenv(CC_JAVA_VERSION, "9")
	defer func() {
		r := recover()
		err, ok := r.(error)
		if !ok || !strings.Contains(err.Error(), CC_JAVA_VERSION) {
			t.Errorf("recovered %v, want an error naming %s", r, CC_JAVA_VERSION)
		}
	}()
	MustGetJavaVersion()
	t.Error("MustGetJavaVersion(9) should have panicked")
}

func TestScopes(t *testing.T) {
	if !slices.Contains(ScopeNode, CC_NODE_BUILD_TOOL) {
		t.Errorf("ScopeNode = %v, want it to hold CC_NODE_BUILD_TOOL", ScopeNode)
	}
	if !slices.Equal(Scopes["node"], ScopeNode) {
		t.Errorf(`Scopes["node"] = %v, want ScopeNode`, Scopes["node"])
	}
	for scope, names := range Scopes {
		if len(names) == 0 || !slices.IsSorted(names) {
			t.Errorf("scope %s: %v should be a non-empty sorted list", scope, names)
		}
	}
}
