package structemplate

import (
	"encoding/json"

	"github.com/drone/envsubst/v2"
	"github.com/pkg/errors"
)

// RenderStrSlotTemplate Rendering a string template containing StrSlot params with the values map.
// @Param tmpl The template string
// @Param valuesMapOfInterface (optional) values of parameters
// @Param valuesMapOfString (optional) value of parameters
// @Return result Rendered string result
// @Return missingKeys missing keys that defined in the template without default value and no value is provided
// @Return err Other errors
func RenderStrSlotTemplate(tmpl string, valuesMapOfInterface map[string]interface{}, valuesMapOfString map[string]string) (result string, missingKeys []string, err error) {
	envTmpl, err := envsubst.Parse(tmpl)
	if err != nil {
		return "", nil, errors.Wrap(err, "cannot parse the template")
	}
	if valuesMapOfInterface == nil {
		valuesMapOfInterface = make(map[string]interface{}, 0)
	}
	if valuesMapOfString == nil {
		valuesMapOfString = make(map[string]string, 0)
	}

	var missingParams []string
	execFunc := func(key string) string {
		v, iok := valuesMapOfInterface[key]
		vs, sok := valuesMapOfString[key]
		if !sok && !iok {
			// missing param
			missingParams = append(missingParams, key)
			return ""
		}

		var valueStr string
		if sok {
			valueStr = vs
		} else {
			switch v := v.(type) {
			case string:
				valueStr = v
			default:
				valueB, err := json.Marshal(v)
				if err != nil {
					missingParams = append(missingParams, key)
					return ""
				}
				valueStr = string(valueB)
			}
		}
		return valueStr
	}

	result, err = envTmpl.Execute(execFunc)
	if err != nil {
		return "", nil, errors.Wrap(err, "cannot render the template")
	}
	return result, missingKeys, nil
}
