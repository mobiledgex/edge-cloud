package edgeproto

import (
	"encoding/json"
	fmt "fmt"
	"sort"

	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
)

type AlertKey string

func (m AlertKey) GetKeyString() string {
	return string(m)
}

func (m *AlertKey) Matches(o *AlertKey) bool {
	return string(*m) == string(*o)
}

func (m AlertKey) NotFoundError() error {
	return fmt.Errorf("Alert key %s not found", m.GetKeyString())
}

func (m AlertKey) ExistsError() error {
	return fmt.Errorf("Alert key %s already exists", m.GetKeyString())
}

func (m AlertKey) GetTags() map[string]string {
	alert := Alert{}
	AlertKeyStringParse(string(m), &alert)
	return alert.Labels
}

func (m *Alert) GetObjKey() objstore.ObjKey {
	return m.GetKey()
}

func (m *Alert) GetKey() *AlertKey {
	key := m.GetKeyVal()
	return &key
}

func (m *Alert) GetKeyVal() AlertKey {
	return AlertKey(MapKey(m.Labels))
}

func (m *Alert) SetKey(key *AlertKey) {
	AlertKeyStringParse(string(*key), m)
}

type elem struct {
	Key string
	Val string
}

// MapKey sorts the elements in the map and generates a json string that can be
// used as a hash table key.
func MapKey(m map[string]string) string {
	elems := make([]elem, 0, len(m))
	for k, v := range m {
		elems = append(elems, elem{Key: k, Val: v})
	}
	sort.Slice(elems, func(i, j int) bool {
		return elems[i].Key < elems[j].Key
	})
	// Order of elements in marshalled output is the same
	// as the order of the elements inserted into the map.
	out := make(map[string]string, len(m))
	for _, e := range elems {
		out[e.Key] = e.Val
	}
	byt, err := json.Marshal(out)
	if err != nil {
		log.FatalLog("Failed to marshal map elements list", "map", m, "err", err)
	}
	return string(byt)
}

func AlertKeyStringParse(str string, obj *Alert) {
	obj.Labels = make(map[string]string)
	err := json.Unmarshal([]byte(str), &obj.Labels)
	if err != nil {
		log.FatalLog("Failed to unmarshal AlertKey key string", "str", str)
	}
}
