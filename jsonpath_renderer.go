package structemplate

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// RenderJsonPathParams 为一个Unstructured对象渲染一组JsonPath param
func RenderJsonPathParams(objsMap map[schema.GroupVersionKind][]*unstructured.Unstructured, paramsDef []TemplateDynamicParam, valuesMap map[string]interface{}) error {
	var err error = nil
	for _, param := range paramsDef {
		if param.ParamType != ParamTypeJsonPath {
			// skip non-JsonPath type params
			continue
		}

		// for every inject target object set value of that json path
		for _, gvkTarget := range param.ValueInjectTargets {
			value, exist := valuesMap[param.ParamCode]
			if !exist {
				value = param.Default
			}
			if value == nil {
				if param.Optional {
					continue
				}
				return errors.New("必填参数缺失:" + param.ParamCode)
			}
			err = RenderObjsWithOneJsonPathParam(objsMap[gvkTarget.TargetGVK], &param, &gvkTarget, value)
			if err != nil {
				break
			}
		}
	}
	return err
}

func RenderObjsWithOneJsonPathParam(objs []*unstructured.Unstructured, paramDef *TemplateDynamicParam, paramPath *JsonPathParamTarget, value interface{}) error {
	var err error = nil
	for _, obj := range objs {
		if err = RenderJsonPathParamForUnstructuredObj(obj, paramDef, paramPath, value); err != nil {
			break
		}
	}
	return err
}

// RenderJsonPathParamForUnstructuredObj 为一个Unstructured Object渲染一个参数，自动识别label selector并过滤
func RenderJsonPathParamForUnstructuredObj(obj *unstructured.Unstructured, paramDef *TemplateDynamicParam, paramPath *JsonPathParamTarget, value interface{}) error {
	// 处理数组元素追加模式
	if paramDef.AppendArray {
		if err := AppendArrayField(obj, paramPath.ParamJsonPath, value); err != nil {
			return err
		}
	}

	// 处理MapKV追加模式
	if len(paramDef.MapKey) > 0 {
		if err := AppendMapForUnstructuredObj(obj, paramPath.ParamJsonPath, paramDef.MapKey, value); err != nil {
			return err
		}
	}

	// 处理一般属性设置模式
	if err := SetValueOfUnstructredObj(obj, paramPath.ParamJsonPath, value); err != nil {
		return err
	}
	return nil
}

func parseKeyPath(keyPath string) []string {
	t1 := strings.Trim(keyPath, "$")
	t1 = strings.Trim(t1, ".")
	return strings.Split(t1, ".")
}

// AppendArrayField 为指定Unstructured对象的数组类型字段增加值
// *将自动判断keyPath指定对象是否为数组类型，或为空时自动创建数组
// *若keyPath不是
func AppendArrayField(obj *unstructured.Unstructured, keyPath string, value interface{}) error {
	kp := parseKeyPath(keyPath)
	subObj, exists, err := unstructured.NestedFieldNoCopy(obj.Object, kp...)
	if err != nil {
		return err
	}

	if !exists {
		// 全新字段，创建数组直接设置
		unstructured.SetNestedField(obj.Object, []interface{}{value}, kp...)
	} else {
		// 检查数据类型为slice
		t := reflect.TypeOf(subObj)
		switch t.Kind() {
		case reflect.Slice, reflect.Array:
			subObj, ok := subObj.([]interface{})
			if !ok {
				return errors.New("转换目标字段为slice出错:" + keyPath)
			}
			newSli := append(subObj, value)

			// 重新替换
			return unstructured.SetNestedSlice(obj.Object, newSli, kp[:len(kp)-1]...)
		default:
			return errors.New("目标字段不是slice或array")
		}
	}
	return nil
}

// SetFieldOfSetValueOfUnstructredObjUnstructured 为指定的Unstructured对象在keyPath指定的位置上设置任意值
// keyPath: `.spec.name1.name2` 格式的json path表达式
// value: 需要设置的任意值
//
//	为Unstructured对象指定属性设置指定value，不支持嵌套map中使用带'.'的key
func SetValueOfUnstructredObj(obj *unstructured.Unstructured, jsonPathKey string, value interface{}) error {
	return AppendMapForUnstructuredObj(obj, jsonPathKey, "", value)
}

// AppendMapForUnstructuredObj 为Unstructured对象指定属性设置指定value，针对map对象添加属性支持带有'.'的key
func AppendMapForUnstructuredObj(obj *unstructured.Unstructured, mapParamJsonPathKey string, key string, value interface{}) error {
	processFunc := func(targetValue interface{}) error {
		// 完成参数设定
		// 解析jsonPath表达式 e.g.  `.metadata.namespace` --> []string{"metadata", "namespace"}
		jsonPathArr := parseKeyPath(mapParamJsonPathKey)
		if len(key) > 0 {
			jsonPathArr = append(jsonPathArr, key)
		}
		return unstructured.SetNestedField(obj.Object, targetValue, jsonPathArr...)
	}

	safeValue := reflect.ValueOf(value)
	switch safeValue.Kind() {
	case reflect.Slice:
		genericSlice := make([]interface{}, safeValue.Len())
		for i := 0; i < safeValue.Len(); i++ {
			genericSlice[i] = safeValue.Index(i).Interface()
		}
		return processFunc(genericSlice)
	case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint8, reflect.Int, reflect.Int16, reflect.Int32, reflect.Int8:
		intV, err := strconv.ParseInt(fmt.Sprintf("%d", value), 10, 64)
		if err != nil {
			return errors.Wrap(err, "整数型参数解析出错")
		}
		return processFunc(intV)
	case reflect.Float32:
		return processFunc(value.(float64))
	default:
		return processFunc(value)
	}
}
