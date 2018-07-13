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

type KVStore interface {
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
	// Put the key-value pair, regardless of whether it already exists or not.
	Put(key, val string) (int64, error)
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

var ErrKVStoreNotInitialized = errors.New("Object Storage not initialized")
var ErrKVStoreKeyNotFound = errors.New("Key not found")
var ErrKVStoreKeyExists = errors.New("Key already exists")

// Any object that wants to be stored in the database
// needs to implement the Obj interface.
type Obj interface {
	// Validate checks all object fields to make sure they do not
	// contain invalid data. Primarily used to validate data passed
	// in by a user. Fields is an array of specified fields for
	// update. It will be nil for create.
	Validate(fields map[string]struct{}) error
	// CopyInFields copies in modified fields for an Update.
	CopyInFields(src Obj)
	// GetKey returns the ObjKey that uniquely identifies the object.
	GetKey() ObjKey
	// HasFields returns true if the object contains a Fields []string
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
}

type SyncCbAction int32

const (
	SyncUpdate SyncCbAction = iota
	SyncDelete
	SyncListStart
	SyncList
	SyncListEnd
)

var SyncActionStrs = [...]string{
	"update",
	"delete",
	"list-start",
	"list",
	"list-end",
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

func DbKeyString(typ string, key ObjKey) string {
	return fmt.Sprintf("%s/%s", DbKeyPrefixString(typ), key.GetKeyString())
}

func DbKeyPrefixString(typ string) string {
	return fmt.Sprintf("%d/%s", GetRegion(), typ)
}

func DbKeyPrefixRemove(key string) string {
	ii := strings.IndexByte(key, '/')
	key = key[ii+1:]
	ii = strings.IndexByte(key, '/')
	return key[ii+1:]
}

func DbKeyPrefixParse(inkey string) (region, typ, key string, err error) {
	key = inkey
	ii := strings.IndexByte(key, '/')
	if ii == -1 {
		return "", "", "", errors.New("No region prefix on db key")
	}
	region = key[:ii]
	key = key[ii+1:]

	ii = strings.IndexByte(key, '/')
	if ii == -1 {
		return "", "", "", errors.New("No type prefix on db key")
	}
	typ = key[:ii]
	key = key[ii+1:]

	return region, typ, key, nil
}
