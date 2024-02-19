package structemplate

import (
	"errors"
	"fmt"
)

// SetNestedField 为一个JSON map结构在指定的属性路径上设置值
// 设置的值进行复制
func SetNestedField(object map[string]interface{}, jsonPath string, value interface{}) error {
	if jsonPath == "" {
		return errors.New("Param jsonPath is empty.")
	}

	paths := strings.split(jsonPath, ".")

	return nil
}

// DeepCopyJSONValue deep copies the passed value, assuming it is a valid JSON representation i.e. only contains
// types produced by json.Unmarshal() and also int64.
// bool, int64, float64, string, []interface{}, map[string]interface{}, json.Number and nil
func DeepCopyJSONValue(x interface{}) interface{} {
	switch x := x.(type) {
	case map[string]interface{}:
		if x == nil {
			// Typed nil - an interface{} that contains a type map[string]interface{} with a value of nil
			return x
		}
		clone := make(map[string]interface{}, len(x))
		for k, v := range x {
			clone[k] = DeepCopyJSONValue(v)
		}
		return clone
	case []interface{}:
		if x == nil {
			// Typed nil - an interface{} that contains a type []interface{} with a value of nil
			return x
		}
		clone := make([]interface{}, len(x))
		for i, v := range x {
			clone[i] = DeepCopyJSONValue(v)
		}
		return clone
	case string, int64, int32, int16, int8, int, bool, rune, float64, float32, nil:
		return x
	default:
		panic(fmt.Errorf("cannot deep copy %T", x))
	}
}
