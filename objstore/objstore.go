package objstore

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// Use for version passed to Update to ignore version check
const ObjStoreUpdateVersionAny int64 = 0

// Callback function for List function
type ListCb func(key, val []byte, rev int64) error

type ObjStore interface {
	// Create creates an object with the given string key and value.
	// Create should fail if the key already exists.
	// It returns the revision (transaction) number and any error.
	Create(key, val string) (int64, error)
	// Update updates an object with the given string key and value.
	// Update should fail if the key does not exist, or if the version
	// doesn't match (meaning some other thread has already updated it).
	// It returns the revision (transaction) number and any error.
	Update(key, val string, version int64) (int64, error)
	// Delete deletes an object with the given key string.
	// It returns the revision (transaction) number and any error.
	Delete(key string) (int64, error)
	// Get retrieves a single object with the given key string.
	// Get returns the data, a version (not revision) number, and any error.
	Get(key string) ([]byte, int64, error)
	// List retrives all objects that have the given key string prefix.
	List(key string, cb ListCb) error
	// Sync is a blocking call used to keep in sync with the database.
	// The initial call (or sometimes after a reconnect), the full set
	// of objects will be called back. Afterwards, or after reconnect
	// if the history is still present, only changes will be called back.
	// It is up to the caller to resync their local cache given actions
	// of SyncAllStart and SyncAllEnd. Any objects that were not received
	// during that time must be removed from the local cache.
	// Use a context with cancel to be able to cancel the call.
	Sync(ctx context.Context, key string, cb SyncCb) error
}

var ErrObjStoreNotInitialized = errors.New("Object Storage not initialized")
var ErrObjStoreKeyNotFound = errors.New("Key not found")
var ErrObjStoreKeyExists = errors.New("Key already exists")

// Any object that wants to be stored in the database
// needs to implement the Obj interface.
type Obj interface {
	// Validate checks all object fields to make sure they do not
	// contain invalid data. Primarily used to validate data passed
	// in by a user.
	Validate() error
	// CopyInFields copies in modified fields for an Update.
	CopyInFields(src Obj)
	// GetKey returns the ObjKey that uniquely identifies the object.
	GetKey() ObjKey
	// HasFields returns true if the object contains a Fields bitmap
	// field used for updating only certain fields on Update.
	HasFields() bool
}

// ObjKey is the struct on the Object that uniquely identifies the Object.
type ObjKey interface {
	// GetKeyString returns a string representation of the ObjKey
	GetKeyString() string
	// Validate checks that the key object fields do not contain
	// invalid or missing data.
	Validate() error
	// TypeString returns a string representing the object type
	// that is used as a prefix when creating the key used to store
	// the object in the database.
	TypeString() string
}

type SyncCbAction int32

const (
	SyncUpdate SyncCbAction = iota
	SyncDelete
	SyncListStart
	SyncList
	SyncListEnd
	SyncRevOnly
)

var SyncActionStrs = [...]string{
	"update",
	"delete",
	"list-start",
	"list",
	"list-end",
	"sync-rev-only",
}

type SyncCb func(data *SyncCbData)

type SyncCbData struct {
	// action on the data
	Action SyncCbAction
	// key - will be nil for action SyncListAllStart/End/Error
	Key []byte
	// value - will be nil for action SyncListAllStart/End/Error
	Value []byte
	// global revision of the data
	Rev int64
}

func DbKeyString(key ObjKey) string {
	return fmt.Sprintf("%s/%s", DbKeyPrefixString(key), key.GetKeyString())
}

func DbKeyPrefixString(key ObjKey) string {
	return fmt.Sprintf("%d/%s", GetRegion(), key.TypeString())
}

func DbKeyPrefixRemove(key string) string {
	ii := strings.IndexByte(key, '/')
	key = key[ii+1:]
	ii = strings.IndexByte(key, '/')
	return key[ii+1:]
}
