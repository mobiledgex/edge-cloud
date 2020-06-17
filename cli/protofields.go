package cli

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"

	yaml "github.com/mobiledgex/yaml/v2"
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
		fval, ok := getProtoTag(sf)
		if !ok {
			continue
		}

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

// Get a map of fields specified by the fields array. This is the opposite of
// GetSpecifiedFields above, converting a proto fields list "[1, 2.2]" to a map
// of obj fields with their associated data from obj.
func GetSpecifiedFieldsData(fields []string, obj interface{}, ns FieldNamespace) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	// It's easiest to use the yaml marshaller to convert to the map
	// and then remove the ones that weren't specified, rather than
	// trying to build the map ourselves from the specified fields.
	// Make sure marshaling does not omit empty.
	buf := bytes.Buffer{}
	enc := yaml.NewEncoder(&buf)
	enc.SetOmitEmpty(false)
	err := enc.Encode(obj)
	if err != nil {
		return data, err
	}
	err = yaml.Unmarshal(buf.Bytes(), &data)
	if err != nil {
		return data, err
	}
	// convert fields to map for easy lookup
	fmap := make(map[string]struct{})
	for _, f := range fields {
		// adds all parent fields as well
		tags := strings.Split(f, ".")
		for ii := len(tags); ii >= 0; ii-- {
			tag := strings.Join(tags[:ii], ".")
			fmap[tag] = struct{}{}
		}
	}
	pruneFields(data, fmap, reflect.TypeOf(obj), ns, []string{})
	return data, nil
}

func pruneFields(data map[string]interface{}, fmap map[string]struct{}, t reflect.Type, ns FieldNamespace, fvals []string) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	for key, val := range data {
		sf, ok := FindField(t, key, ns)
		if !ok {
			continue
		}
		fval, ok := getProtoTag(sf)
		if !ok {
			continue
		}
		fstr := strings.Join(append(fvals, fval), ".")
		if _, found := fmap[fstr]; !found {
			delete(data, key)
			continue
		}
		if subdata, ok := val.(map[string]interface{}); ok {
			// sub struct
			pruneFields(subdata, fmap, sf.Type, ns, append(fvals, fval))
		}
	}
}

func getProtoTag(sf reflect.StructField) (string, bool) {
	tag := sf.Tag.Get("protobuf")
	tagvals := strings.Split(tag, ",")
	if len(tagvals) < 2 {
		return "", false
	}
	return tagvals[1], true
}

func GetGenericObj(dataMap interface{}) (map[string]interface{}, error) {
	if dataMap == nil {
		return nil, fmt.Errorf("nil dataMap")
	}
	if m, ok := dataMap.(map[string]interface{}); ok {
		return m, nil
	}
	return nil, fmt.Errorf("expected map[string]interface{} but was %T", dataMap)
}

func GetGenericObjFromList(dataMap interface{}, idx int) (map[string]interface{}, error) {
	if dataMap == nil {
		return nil, fmt.Errorf("nil dataMap")
	}
	if list, ok := dataMap.([]interface{}); ok {
		if obj, ok := list[idx].(map[string]interface{}); ok {
			return obj, nil
		}
		return nil, fmt.Errorf("index %d expected map[string]interface{} but was %T", idx, list[idx])
	}
	return nil, fmt.Errorf("expected []interface{} but was %T", dataMap)
}
