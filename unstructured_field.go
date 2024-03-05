package structemplate

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

func ParseKeyPath(keyPath string) []string {
	t1 := strings.Trim(keyPath, "$")
	t1 = strings.Trim(t1, ".")
	return strings.Split(t1, ".")
}

// GetValueOfNestedField gets the value of field specified by `jsonPath` from the target object.
func GetValueOfNestedField(object map[string]interface{}, jsonPath string) (interface{}, error) {
	if len(jsonPath) < 1 || jsonPath == "." {
		return DeepCopyJSONValue(object), nil
	}

	paths := ParseKeyPath(jsonPath)
	var field interface{} = DeepCopyJSONValue(object)

	var tracedPath string = ""
	for i := 0; i < len(paths); i++ {
		key := paths[i]
		tracedPath += "." + key
		switch reflect.TypeOf(field).Kind() {
		case reflect.Slice:
			idx, err := ParseJsonPathArrayIndex(key)
			if err != nil {
				return nil, errors.Wrap(err, "cannot parse the index of array field: "+tracedPath)
			}
			tmpField := field.([]interface{})
			if len(tmpField) <= int(idx) {
				return nil, errors.New(fmt.Sprintf("array field index out of bounds: %d of %s", idx, tracedPath))
			}
			field = tmpField[idx]
		case reflect.Map:
			tmpField := field.(map[string]interface{})
			field = tmpField[key]
		default:
			return nil, errors.New("field does not exist: " + tracedPath)
		}
	}
	return field, nil
}

// SetNestedField sets a value in the structure of object
// @Param appendArray: the element to operating is an array and the value should be appended in the array
// @Param value: the value to inject. When append array is true, and value is an array, the elements in `value` will all be appended to the template.
func SetNestedField(object map[string]interface{}, jsonPath string, value interface{}, appendArray bool) error {
	if jsonPath == "" {
		return errors.New("param jsonPath is empty")
	}

	paths := ParseKeyPath(jsonPath)
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
				return errors.Wrap(err, fmt.Sprintf("arror parse array index: %s", currentKey))
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
			jumperBackNode = jumper
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
				} //if: end processing missing nodes

				jumper = jumperHolder[currentKey]
			} else {
				// jump to next node
				jumper = nextJumper
			}
		default:
			return fmt.Errorf("element cannot be indexed: %s", currentKey)
		}
	} // end finding last node

	// processing last key and set the value
	lastKey := paths[len(paths)-1]

	// when append array is set, the jumper must be a map
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

		switch reflect.TypeOf(value).Kind() {
		case reflect.Slice:
			sliceValue := reflect.ValueOf(value)
			newSlice := jumperObj[lastKey].([]interface{})
			for i := 0; i < sliceValue.Len(); i++ {
				newSlice = append(newSlice, sliceValue.Index(i).Interface())
			}
			jumperObj[lastKey] = newSlice
		default:
			jumperObj[lastKey] = append(targetArr, value)
		}

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
	if !matched {
		return -1, errors.New("parse index expression failed")
	}
	if err != nil {
		return -1, errors.Wrap(err, "parse index expression failed")
	}

	idx := strings.TrimLeft(idxExp, "[")
	idx = strings.TrimRight(idx, "]")
	idxNum, err := strconv.Atoi(idx)
	if err != nil {
		return -1, errors.Wrap(err, "illegal index number: "+idx)
	}
	if idxNum < 0 {
		return -1, errors.New("index less than 0")
	}
	return int64(idxNum), nil
}

// DeepCopyJSONValue deep copies the passed value, assuming it is a valid JSON representation i.e. only contains
// types produced by json.Unmarshal() and also int64.
// bool, int64, float64, string, []interface{}, map[string]interface{}, json.Number and nil
func deepCopyJSONValuebackup(x interface{}) interface{} {
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

func DeepCopyJSONValue(x interface{}) interface{} {
	val := reflect.ValueOf(x)
	switch val.Kind() {
	case reflect.Map:
		if val.IsNil() {
			return x
		}
		clone := reflect.MakeMap(val.Type())
		for _, k := range val.MapKeys() {
			clone.SetMapIndex(k, reflect.ValueOf(DeepCopyJSONValue(val.MapIndex(k).Interface())))
		}
		return clone.Interface()
	case reflect.Slice:
		if val.IsNil() {
			return x
		}
		clone := reflect.MakeSlice(val.Type(), val.Len(), val.Cap())
		for i := 0; i < val.Len(); i++ {
			clone.Index(i).Set(reflect.ValueOf(DeepCopyJSONValue(val.Index(i).Interface())))
		}
		return clone.Interface()
	case reflect.String, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8, reflect.Int, reflect.Bool, reflect.Float64, reflect.Float32:
		return x
	default:
		if val.IsValid() && val.CanInterface() {
			return x
		}
		panic(fmt.Errorf("cannot deep copy %T", x))
	}
}
