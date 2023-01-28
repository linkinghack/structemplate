# Structemplate
A structured data template rendering engine for K8s objects
that supports defining flexible dynamic parameters to modify arbitrary fields in the template.

## Structured data template
Structured data simply means some data structured with a common serializing method like `json` or `yaml`.

In Kubernetes manifests, there are often some fields that needs be specified dynamically according to different
scenarios when most of them don't need to change.

Reusing the manifests can increase the efficiency of cloud-native deployments, which is `Helm` is doing.

`Structemplate` is different from `Helm`. It uses a precise way to define the variational parts of a manifest template
and use values list to render a template afterwards.

## Dynamic Parameter
A dynamic parameter is a well-defined variable in the structured data template.

For example:
```yaml
name: ${OBJ_NAME:=Job}
annotations:
  foo: bar
```
the variable `OBJ_NAME` is a dynamic parameter that can be replaced with a specific value while rendering.
There is also a default value defined in the template which can be used while rendering if the value of this variable is not provided.

More complicated situation is about dynamic maps and arrays in the template which can be located and assigned with `json path`. 

We can use a well-defined dynamic parameter to manage the variables and the rendering procedure in a system.

## Usage
For general string templates (not only yaml/json or k8s manifests) StrSlots parameters can be used.

For K8s manifests StrSlot and JsonPath params can be used.

`go get github.com/linkinghack/structemplate`