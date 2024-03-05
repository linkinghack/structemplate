package structemplate

import (
	"bytes"
	"encoding/json"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
)

var manifest string = `
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: TLSRoute
metadata:
  name: test-target-tlsroute
  namespace: gateways
  labels: 
    istio: test-target-gateway
spec:
  hostnames:
  - "hostname1.example.com"
  - "hostname2.example.com"
  parentRefs:
  - group: gateway.networking.k8s.io
    kind: Gateway
    name: example-gateway
    namespace: gateways
    port: 20022
  rules:
  - backendRefs:
    - name: test-target
      namespace: gateways
      port: 6443
`

var param = TemplateDynamicParam{
	ParamCode:     "TLS_SNI_HOSTS",
	ParamName:     "TLS SNI hostname",
	FunctionScope: "SYSTEM",
	ParamType:     ParamTypeJsonPath,
	ValueInjectTargets: []JsonPathParamTarget{
		{
			TargetGVK:     schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1alpha2", Kind: "TLSRoute"},
			ParamJsonPath: ".spec.hostnames",
		},
	},
	Optional:      false,
	Default:       "default.example.com",
	ValueDataType: "string",
	AppendArray:   true,
}

func TestRenderJsonpathParam_AppendArray(t *testing.T) {
	obj := parseObject(t)
	unstructuredObj := unstructured.Unstructured{Object: obj}
	var newParam = param

	objJson, _ := json.Marshal(obj)
	t.Logf("Before append array, the object: %s\n", objJson)

	err := RenderJsonPathParamForUnstructuredObj(&unstructuredObj, &newParam, &newParam.ValueInjectTargets[0], "correct.sni-hostname.example.com")
	if err != nil {
		t.Logf("Failed render JsonPath param: %+v", err)
		t.FailNow()
		return
	}

	objJson, _ = json.Marshal(obj)
	t.Logf("After append array: %s \n", string(objJson))

	// check result
	hostnames, err := GetValueOfNestedField(unstructuredObj.Object, ".spec.hostnames")
	if err != nil {
		t.Error(err)
		t.FailNow()
		return
	}
	hostnamesArray := hostnames.([]interface{})
	t.Logf("hostnames array: %+v", hostnames)
	if len(hostnamesArray) != 3 || hostnamesArray[2] != "correct.sni-hostname.example.com" {
		t.Log("Append array test failed")
		t.FailNow()
		return
	}

	// show results

	jsonResult, err := json.Marshal(obj)
	if err != nil {
		t.Log(err)
		t.FailNow()
		return
	}
	t.Logf("After render JsonPath param: %s", jsonResult)
}

func TestRenderJsonpathParam_ArrayAddObj(t *testing.T) {
	obj := parseObject(t)
	unstructuredObj := unstructured.Unstructured{Object: obj}
	var newParam = param

	newParam.AppendArray = false
	newParam.ValueInjectTargets[0].ParamJsonPath = ".spec.hostnames.fakeKey"
	err := RenderJsonPathParamForUnstructuredObj(&unstructuredObj, &newParam, &newParam.ValueInjectTargets[0], "fakeValue")
	if err == nil {
		t.Log("Expected error does not occurred")
		t.FailNow()
		return
	}
	t.Logf("illegal json path error: %s", err.Error())
}

func TestRenderJsonpathParam_AppendMap(t *testing.T) {
	var newParam = param
	newParam.MapKey = "newObjectKey"
	newParam.AppendArray = false
	newParam.ValueInjectTargets[0].ParamJsonPath = ".spec"
	obj := parseObject(t)
	unstructuredObj := unstructured.Unstructured{Object: obj}
	RenderJsonPathParamForUnstructuredObj(&unstructuredObj, &newParam, &newParam.ValueInjectTargets[0], "newObjectValue")
	objJson, _ := json.Marshal(obj)
	t.Logf("After append map: %s\n", string(objJson))
}

func TestRenderJsonpathParam_NormalSetField(t *testing.T) {
	var newParam = param
	newParam.AppendArray = false
	newParam.ValueInjectTargets[0].ParamJsonPath = ".spec.newEmptyObject.newEmptyField"
	obj := parseObject(t)
	unstructuredObj := unstructured.Unstructured{Object: obj}
	RenderJsonPathParamForUnstructuredObj(&unstructuredObj, &newParam, &newParam.ValueInjectTargets[0], "newObjectValue")
	objJson, _ := json.Marshal(obj)
	t.Logf("After normal set: %s\n", string(objJson))
}

func parseObject(t *testing.T) map[string]interface{} {
	yamlReader := bytes.NewReader([]byte(manifest))
	decoder := yaml.NewYAMLOrJSONDecoder(yamlReader, 0)
	obj := make(map[string]interface{})
	if err := decoder.Decode(&obj); err != nil {
		t.Log(err)
		t.FailNow()
		return nil
	}
	return obj
}

func TestJsonpathGetterSetter(t *testing.T) {
	obj := parseObject(t)
	SetNestedField(obj, ".spec.parentRefs.[0].name", "JSONHACKED", false)

	v, err := GetValueOfNestedField(obj, ".spec.parentRefs.[0].name")
	if err != nil {
		t.Logf("Failed get value: %+v", err)
		t.FailNow()
		return
	}
	if v != "JSONHACKED" {
		t.Logf("Value retrived: %+v", v)
		t.FailNow()
		return
	}
	t.Logf("Retrived: %s", v)

	jsonResult, err := json.Marshal(obj)
	if err != nil {
		t.Log(err)
		t.FailNow()
		return
	}
	t.Logf("After modify: %+v", string(jsonResult))
}
