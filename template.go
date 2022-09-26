package sdk

import (
	"regexp"
)

// Tpl templatise an URL
// Ex: /test/{A} with A=>"id" result /test/id
func Tpl(url string, vars map[string]string) string {
	for key, val := range vars {

		reg, err := regexp.Compile("{" + key + "}")
		if err != nil {
			return ""
		}

		url = reg.ReplaceAllString(url, val)
	}

	return url
}

func WithQuery(url string, params []Parameter) string {
	return url
}
