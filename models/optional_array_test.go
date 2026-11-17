package models

import (
	"encoding/json"
	"testing"
)

// The API tells an absent array (keep the current tags) from an empty one
// (remove them all): the payload has to be able to say both.
func TestOptionalArrayAbsentOrEmpty(t *testing.T) {
	for name, tc := range map[string]struct {
		tags []SemanticTag
		want string
	}{
		"nil is left out": {nil, `{}`},
		"empty is sent":   {[]SemanticTag{}, `{"tags":[]}`},
		"values are sent": {[]SemanticTag{"a"}, `{"tags":["a"]}`},
	} {
		t.Run(name, func(t *testing.T) {
			got, err := json.Marshal(ClusterPatchPayload{Tags: tc.tags})
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tc.want {
				t.Errorf("got %s, want %s", got, tc.want)
			}
		})
	}
}
