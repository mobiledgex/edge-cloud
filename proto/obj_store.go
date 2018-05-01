package proto

import "errors"

// Use for version passed to Update to ignore version check
const ObjStoreUpdateVersionAny int64 = 0

// Callback function for List function
type ListCb func(key, val []byte) error

type ObjStore interface {
	Create(key, val string) error
	Update(key, val string, version int64) error
	Delete(key string) error
	Get(key string) ([]byte, int64, error)
	List(key string, cb ListCb) error
}

var ObjStoreErrNotInitialized = errors.New("Object Storage not initialized")
var ObjStoreErrKeyNotFound = errors.New("Key not found")
var ObjStoreErrKeyExists = errors.New("Key already exists")
