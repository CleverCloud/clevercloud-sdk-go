package kubernetes

import "testing"

// A query array repeats its key once per value (OpenAPI's default `form` + `explode`).
func TestQueryArrayRepeatsItsKey(t *testing.T) {
	got := buildQueryString(WithStatus([]string{"ACTIVE", "FAILED"}), WithTypeparam("admin"))
	if want := "status=ACTIVE&status=FAILED&type=admin"; got != want {
		t.Errorf("buildQueryString = %q, want %q", got, want)
	}
}
