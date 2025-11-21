package utils

import (
	"testing"
)

func TestPath(t *testing.T) {
	tests := []struct {
		name     string
		template string
		params   []any
		expected string
	}{
		{
			name:     "no parameters",
			template: "/v4/health",
			params:   nil,
			expected: "/v4/health",
		},
		{
			name:     "string parameter",
			template: "/v4/addon-providers/addon-pulsar/addons/%s",
			params:   []any{"test-addon"},
			expected: "/v4/addon-providers/addon-pulsar/addons/test-addon",
		},
		{
			name:     "string parameter with special chars",
			template: "/v4/addon-providers/addon-pulsar/addons/%s",
			params:   []any{"test addon/with spaces"},
			expected: "/v4/addon-providers/addon-pulsar/addons/test%20addon%2Fwith%20spaces",
		},
		{
			name:     "multiple parameters",
			template: "/v4/addon-providers/addon-pulsar/addons/%s/persistent-topics/%s",
			params:   []any{"addon-id", "topic-name"},
			expected: "/v4/addon-providers/addon-pulsar/addons/addon-id/persistent-topics/topic-name",
		},
		{
			name:     "integer parameter",
			template: "/v4/outbox/records/from/%s/limit/%s",
			params:   []any{123, 50},
			expected: "/v4/outbox/records/from/123/limit/50",
		},
		{
			name:     "int64 parameter",
			template: "/v4/outbox/records/from/%s",
			params:   []any{int64(9223372036854775807)},
			expected: "/v4/outbox/records/from/9223372036854775807",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Path(tt.template, tt.params...)
			if result != tt.expected {
				t.Errorf("Path() = %v, want %v", result, tt.expected)
			}
		})
	}
}
