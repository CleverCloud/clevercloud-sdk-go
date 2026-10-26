package products

import "testing"

// The query string carries the parameter under the name the API reads. `for` is a Go
// keyword, so the option is `WithForparam`, but the wire name must stay `for`.
func TestQueryParameterKeepsItsSpecName(t *testing.T) {
	if got := buildQueryString(WithForparam("orga_x")); got != "for=orga_x" {
		t.Errorf("buildQueryString = %q, want %q", got, "for=orga_x")
	}
}
