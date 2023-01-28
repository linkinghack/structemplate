package structemplate

import (
	"github.com/drone/envsubst/v2"
	"github.com/pkg/errors"
)

// RenderStrSlotTemplate Rendering a string template containing StrSlot params with the values map.
// @Param tmpl The template string
// @Param valuesMap The values of parameters
// @Return result Rendered string result
// @Return missingKeys missing keys that defined in the template without default value and no value is provided
// @Return err Other errors
func RenderStrSlotTemplate(tmpl string, valuesMap map[string]string) (result string, missingKeys []string, err error) {
	envTmpl, err := envsubst.Parse(tmpl)
	if err != nil {
		return "", nil, errors.Wrap(err, "cannot parse the template")
	}

	missingParams := []string{}
	execFunc := func(key string) string {
		v, ok := valuesMap[key]
		if !ok {
			// missing param
			missingParams = append(missingParams, key)
			return ""
		}
		return v
	}
	result, err = envTmpl.Execute(execFunc)
	if err != nil {
		return "", nil, errors.Wrap(err, "cannot render the template")
	}

	return result, missingKeys, nil
}
