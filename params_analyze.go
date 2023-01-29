package structemplate

import "log"

// SplitParamsByType splits an array of TemplateDynamicParams into two groups
// by their type (StrSlot or JsonPath) and builds a params map with ParamCodes as keys.
// Returns: StrSlot Params, JsonPath Params, Params Map
func SplitParamsByType(params []TemplateDynamicParam) ([]*TemplateDynamicParam, []*TemplateDynamicParam, map[string]*TemplateDynamicParam) {
	strSlotParams := make([]*TemplateDynamicParam, 1)
	jsonPathParams := make([]*TemplateDynamicParam, 1)
	paramsMap := make(map[string]*TemplateDynamicParam)
	for _, p := range params {
		switch p.ParamType {
		case ParamTypeStrSlot:
			strSlotParams = append(strSlotParams, &p)
		case ParamTypeJsonPath:
			jsonPathParams = append(jsonPathParams, &p)
		default:
			log.Println("Unknown Param type: " + p.ParamType)
		}

		// Add to params map
		paramsMap[p.ParamCode] = &p
	}
	return strSlotParams, jsonPathParams, paramsMap
}
