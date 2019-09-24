package cli

import (
	"reflect"
	"strings"
)

// Get a list of fields specified as their hierarchical id space.
// I.e. a field will be "1", or "2.2", based on the protobuf id and
// hierarchy. Data contains the specified data hierarchically
// arranged. Obj is the protobuf object with protobuf tags on fields
// that corresponds to the data.
func GetSpecifiedFields(data map[string]interface{}, obj interface{}, ns FieldNamespace) []string {
	return getFields(data, reflect.TypeOf(obj), ns, []string{})
}

func getFields(data map[string]interface{}, t reflect.Type, ns FieldNamespace, fvals []string) []string {
	fields := []string{}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	for key, val := range data {
		sf, ok := FindField(t, key, ns)
		if !ok {
			continue
		}
		tag := sf.Tag.Get("protobuf")
		tagvals := strings.Split(tag, ",")
		if len(tagvals) < 2 {
			continue
		}
		fval := tagvals[1]

		if subdata, ok := val.(map[string]interface{}); ok {
			// sub struct
			subfields := getFields(subdata, sf.Type, ns, append(fvals, fval))
			fields = append(fields, subfields...)
			continue
		}
		fstr := strings.Join(append(fvals, fval), ".")
		fields = append(fields, fstr)
	}
	return fields
}
