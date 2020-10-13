package edgeproto

import (
	fmt "fmt"

	"github.com/mobiledgex/edge-cloud/objstore"
)

var OperatorCodeKeyTag = "operatorcode"

type OperatorCodeKey string

func (m OperatorCodeKey) GetKeyString() string {
	return string(m)
}

func (m *OperatorCodeKey) Matches(o *OperatorCodeKey) bool {
	return string(*m) == string(*o)
}

func (m OperatorCodeKey) NotFoundError() error {
	return fmt.Errorf("OperatorCode key %s not found", m.GetKeyString())
}

func (m OperatorCodeKey) ExistsError() error {
	return fmt.Errorf("OperatorCode key %s already exists", m.GetKeyString())
}

func (m OperatorCodeKey) GetTags() map[string]string {
	return map[string]string{
		OperatorCodeKeyTag: string(m),
	}
}

func (m *OperatorCode) GetObjKey() objstore.ObjKey {
	return m.GetKey()
}

func (m *OperatorCode) GetKey() *OperatorCodeKey {
	key := m.GetKeyVal()
	return &key
}

func (m *OperatorCode) GetKeyVal() OperatorCodeKey {
	return OperatorCodeKey(m.Code)
}

func (m *OperatorCode) SetKey(key *OperatorCodeKey) {
	m.Code = string(*key)
}

func CmpSortOperatorCode(a OperatorCode, b OperatorCode) bool {
	return a.GetKey().GetKeyString() < b.GetKey().GetKeyString()
}

func OperatorCodeKeyStringParse(str string, obj *OperatorCode) {
	obj.Code = str
}
