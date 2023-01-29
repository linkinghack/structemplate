package structemplate

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// TemplateDynamicParam Defines a dynamic param that referenced by param code in a template.
// Param values can be referenced by dynamic param definition for multiple times within one template.
// Multiple dynamic params of different param type may be defined with same ParamCode to reference same param value.
// Value for StrSlot param must be of type string while value for JsonPath must have correct data type (interface{} to execute yaml/json marshal).
// StrSlot renderer should accept interface{} value.
type TemplateDynamicParam struct {
	ParamCode string `json:"paramCode"` // 参数唯一标识，用于模板中引用一个确定的值
	ParamName string `json:"paramName"` // 用户可读的变量名称，易于分辨变量功能
	Brief     string `json:"brief"`     // 参数解释

	FunctionScope      string                `json:"functionScope"` // 作用范围 设定系统参数或用户可自定义
	ParamType          string                `json:"paramType"`     // StrSlot, JsonPath  支持两种动态参数设置方式. 基于字符串替换的StrSlot和自定义JsonPath
	ValueInjectTargets []JsonPathParamTarget `json:"valueInjectTargets"`
	Optional           bool                  `json:"optional"` // 是否为可选参数
	Default            interface{}           `json:"default"`

	AvailableOptions []interface{} `json:"availableOptions"` // 预设可选值
	Customizable     bool          `json:"customizable"`     // 是否允许用户自定义。为false时仅支持设定AvailableOptions中预设的值
	ValueDataType    string        `json:"dataType"`         // int, string, float, boolean, object, array[string] 当前仅用于类型提示

	// 对于jsonPath类型参数，处理对象和数组的方式
	AppendArray bool   `json:"appendArray"` // 当JsonPath指向一个数组类型时, 进行替换还是追加
	MapKey      string `json:"mapKey"`      // 当JsonPath指向目标为Map类型时，将在此map中增加一个KV对，此值不为空时表示中增加的KV对中的key
}

type JsonPathParamTarget struct {
	TargetGVK           schema.GroupVersionKind `json:"targetGVK,omitempty"`           // 对于JsonPath类型参数，指定要设置的目标模板对象, 若存在多个同种对象,需要增加label来标识
	ParamJsonPath       string                  `json:"paramJsonPath,omitempty"`       // .param1.param-sub1
	ObjectLabelSelector map[string]string       `json:"objectDistinctLabel,omitempty"` // 用于区分同一个模板中同一种GVK定义的多个不同对象
}

const (
	ParamTypeStrSlot  = "StrSlot"
	ParamTypeJsonPath = "JsonPath"
)
