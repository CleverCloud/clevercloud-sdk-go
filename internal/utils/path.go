package utils

import (
	"fmt"
	"net/url"
	"github.com/miton18/helper/slices"
)

// Path constructs an API path with URL-encoded parameters.
// The first parameter is the path template (with %s placeholders),
// followed by path parameter values that will be URL-encoded.
//
// Example:
//
//	Path("/v4/addon-providers/addon-pulsar/addons/%s", addonId)
//	Path("/v4/addon-providers/addon-pulsar/addons/%s/persistent-topics/%s", addonId, topicName)
func Path(template string, params ...any) string {
	if len(params) == 0 {
		return template
	}

	// URL encode all parameters
	encodedParams := slices.Map(params, func(param any) any {
		return url.PathEscape(fmt.Sprintf("%v", param))
	})

	return fmt.Sprintf(template, encodedParams...)
}
