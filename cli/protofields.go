package cli

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// Get a list of fields specified as their hierarchical id space.
// I.e. a field will be "1", or "2.2", based on the protobuf id and
// hierarchy. Data contains the specified data hierarchically
// arranged. Obj is the protobuf object with protobuf tags on fields
// that corresponds to the data.
func GetSpecifiedFields(data *MapData, obj interface{}) []string {
	return getFields(data.Data, reflect.TypeOf(obj), data.Namespace, []string{})
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

		if subdata, ok := val.(map[string]interface{}); ok && len(subdata) > 0 {
			// sub struct
			subfields := getFields(subdata, sf.Type, ns, append(fvals, fval))
			fields = append(fields, subfields...)
			continue
		}
		if subdataArr, ok := arrayOfMaps(val); ok && len(subdataArr) > 0 {
			// array of sub structs
			if sf.Type.Kind() != reflect.Slice {
				continue
			}
			fields = append(fields, strings.Join(append(fvals, fval), "."))
			// use map to eliminate duplicate fields
			subfieldsMap := make(map[string]struct{})
			for _, subdata := range subdataArr {
				subfields := getFields(subdata, sf.Type.Elem(), ns, append(fvals, fval))
				for _, s := range subfields {
					subfieldsMap[s] = struct{}{}
				}
			}
			subfields := []string{}
			for k, _ := range subfieldsMap {
				subfields = append(subfields, k)
			}
			sort.Strings(subfields)
			fields = append(fields, subfields...)
			continue
		}
		fstr := strings.Join(append(fvals, fval), ".")
		fields = append(fields, fstr)
	}
	return fields
}

// Convert the value to an array of map[string]interface{},
// which corresponds to an array of structs.
func arrayOfMaps(val interface{}) ([]map[string]interface{}, bool) {
	// Val could be a []map[string]interface{} object, or
	// it could be an []interface{} array, where each interface{}
	// is a map[string]interface{} object (an extra level of indirection).
	// The latter case happens when umarshaling yaml into
	// a generic map[string]interface{} data, when the yaml
	// has a field which is an array of substructs.
	if arr, ok := val.([]map[string]interface{}); ok {
		return arr, ok
	}
	if arr, ok := val.([]interface{}); ok {
		mm := make([]map[string]interface{}, len(arr), len(arr))
		for ii, obj := range arr {
			subm, ok := obj.(map[string]interface{})
			if ok {
				mm[ii] = subm
			} else {
				return nil, false
			}
		}
		return mm, true
	}
	return nil, false
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
