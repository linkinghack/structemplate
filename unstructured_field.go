package structemplate

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// SetNestedField set a value in the structure of object
// @Param appendArray: the element to operating is an array and the value should be appended in the array
func SetNestedField(object map[string]interface{}, jsonPath string, value interface{}, appendArray bool) error {
	if jsonPath == "" {
		return errors.New("param jsonPath is empty")
	}

	jsonPath = strings.Trim(jsonPath, ".")
	paths := strings.Split(jsonPath, ".")
	var jumper interface{} = object
	var jumperBackNode interface{} = object

	// finding the node before last key
	for i := 0; i < len(paths)-1; i++ {
		currentKey := paths[i]
		if len(currentKey) < 1 {
			// ignore empty slots
			continue
		}

		switch reflect.TypeOf(jumper).Kind() {
		case reflect.Slice:
			idx, err := ParseJsonPathArrayIndex(currentKey)
			if err != nil {
				return errors.Join(fmt.Errorf("arror parse array index: %s", currentKey), err)
			}
			jumperHolder := jumper.([]interface{})
			if int(idx) >= len(jumperHolder) {
				return fmt.Errorf("array index out of bounds: %s", currentKey)
			}
			if jumperHolder[idx] == nil {
				jumperHolder[idx] = make(map[string]interface{})
			}
			jumperBackNode = jumper
			jumper = jumperHolder[idx]
		case reflect.Map:
			jumperHolder := jumper.(map[string]interface{})
			nextJumper, ok := jumperHolder[currentKey]
			if !ok {
				// Auto create missing nodes
				nextKey := paths[i+1]
				idx, err := ParseJsonPathArrayIndex(nextKey)
				if err == nil && idx > -1 {
					// 1. .missing_property.[idx];  add an array element
					jumperHolder[currentKey] = make([]interface{}, idx+1)
				} else {
					// 2. .missing_property.other_property; add an object element
					jumperHolder[currentKey] = make(map[string]interface{})
				}
				//if: end processing missing nodes
			} else {
				// jump to next node
				jumperBackNode = jumper
				jumper = nextJumper
			}
		default:
			return fmt.Errorf("element cannot be indexed: %s", currentKey)
		}
	} // end finding last node

	lastKey := paths[len(paths)-1]

	// when append array is set, the jumper must be an map
	if appendArray {
		jumperObj, ok := jumper.(map[string]interface{})
		if !ok {
			return fmt.Errorf("append array failed because the node before last node is not an object")
		}

		if jumperObj[lastKey] == nil {
			jumperObj[lastKey] = make([]interface{}, 0)
		}

		lastNodeType := reflect.TypeOf(jumperObj[lastKey]).Kind()
		if lastNodeType != reflect.Slice {
			return fmt.Errorf("append array failed because last node is not an array: %s", lastNodeType.String())
		}

		targetArr := jumperObj[lastKey].([]interface{})
		jumperObj[lastKey] = append(targetArr, value)
		return nil
	}

	// set last node

	jumperType := reflect.TypeOf(jumper).Kind()
	if jumperType == reflect.Slice {
		// 1. fixed index array
		idx, err := ParseJsonPathArrayIndex(lastKey)
		if err != nil {
			return fmt.Errorf("cannot set object property inside an array")
		}
		jumperArray := jumper.([]interface{})
		backNode := jumperBackNode.(map[string]interface{})
		if len(jumperArray) < int(idx+1) {
			return fmt.Errorf("cannot modify the target array: index out of bounds: %s", lastKey)
		}
		jumperArray[idx] = value
		backNode[paths[len(paths)-2]] = jumperArray
		return nil
	}
	// 2. normal object
	jumberObj, ok := jumper.(map[string]interface{})
	if !ok {
		return fmt.Errorf("element cannot be indexed: %s", strings.Join(paths[0:len(paths)-1], "."))
	}
	jumberObj[lastKey] = value
	return nil
}

func ParseJsonPathArrayIndex(idxExp string) (int64, error) {
	matched, err := regexp.Match("^\\[\\d+\\]$", []byte(idxExp))
	if !matched || err != nil {
		return -1, errors.Join(errors.New("Parse index expression failed"), err)
	}

	idx := strings.TrimLeft(idxExp, "[")
	idx = strings.TrimRight(idx, "]")
	idxNum, err := strconv.Atoi(idx)
	if err != nil {
		return -1, errors.Join(errors.New(""), err)
	}
	if idxNum < 0 {
		return -1, errors.New("Index less than 0")
	}
	return int64(idxNum), nil
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
	case string, int64, int32, int16, int8, int, bool, float64, float32, nil: // rune is int32
		return x
	default:
		panic(fmt.Errorf("cannot deep copy %T", x))
	}
}
