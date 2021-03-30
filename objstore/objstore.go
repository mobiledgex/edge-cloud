package objstore

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/coreos/etcd/clientv3/concurrency"
)

// Use for version passed to Update to ignore version check
const ObjStoreUpdateVersionAny int64 = 0

// Callback function for List function
type ListCb func(key, val []byte, rev, modRev int64) error

type KVStore interface {
	// Create creates an object with the given string key and value.
	// Create should fail if the key already exists.
	// It returns the revision (transaction) number and any error.
	Create(ctx context.Context, key, val string) (int64, error)
	// Update updates an object with the given string key and value.
	// Update should fail if the key does not exist, or if the version
	// doesn't match (meaning some other thread has already updated it).
	// It returns the revision (transaction) number and any error.
	Update(ctx context.Context, key, val string, version int64) (int64, error)
	// Delete deletes an object with the given key string.
	// It returns the revision (transaction) number and any error.
	Delete(ctx context.Context, key string) (int64, error)
	// Get retrieves a single object with the given key string.
	// Get returns the data, a version, mod revision, and any error.
	Get(key string, opts ...KVOp) ([]byte, int64, int64, error)
	// Put the key-value pair, regardless of whether it already exists or not.
	Put(ctx context.Context, key, val string, opts ...KVOp) (int64, error)
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
	// ApplySTM applies a Software Transaction Model which basically
	// collects gets/puts and does an all-or-nothing transaction.
	// It tracks revisions for all gets and puts. If any keys were
	// changed before the transaction commits, all changes are aborted.
	// Apply func is the function to make the changes which uses the
	// STM to make the changes.
	// Unfortunately the way etcd sets this up, there's no way to wrap
	// the STM with an objstore-specific interface, so we're stuck exactly
	// implementing the etcd-specific interface.
	// Important: Etcd apparently does not honor the order in which
	// multiple puts appear within an apply func, at least for watch
	// callbacks (perhaps because they all have the same revision ID).
	// If ordering is important, do not use multiple puts in the same STM.
	ApplySTM(ctx context.Context, apply func(concurrency.STM) error) (int64, error)
	// Leases work like etcd leases. A key committed with a lease will
	// automatically be deleted once the lease expires.
	// To avoid that, the KeepAlive call must remain active.
	// Grant creates a new lease
	Grant(ctx context.Context, ttl int64) (int64, error)
	// Revoke a lease
	Revoke(ctx context.Context, lease int64) error
	// KeepAlive keeps a lease alive. This call blocks.
	KeepAlive(ctx context.Context, leaseID int64) error
}

var ErrKVStoreNotInitialized = errors.New("Object Storage not initialized")

// Any object that wants to be stored in the database
// needs to implement the Obj interface.
type Obj interface {
	// Validate checks all object fields to make sure they do not
	// contain invalid data. Primarily used to validate data passed
	// in by a user. Fields is an array of specified fields for
	// update. It will be nil for create.
	Validate(fields map[string]struct{}) error
	// CopyInFields copies in modified fields for an Update.
	//CopyInFields(src Obj)
	// GetObjKey returns the ObjKey that uniquely identifies the object.
	GetObjKey() ObjKey
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
	ValidateKey() error
	// NotFoundError returns a not found error describing the key.
	NotFoundError() error
	// ExistsError returns an already exists error describing the key.
	ExistsError() error
	// Get key tags for logging and tagging
	GetTags() map[string]string
}

type KVOptions struct {
	LeaseID int64
	Rev     int64
}

type KVOp func(opts *KVOptions)

func WithLease(leaseID int64) KVOp {
	return func(opts *KVOptions) { opts.LeaseID = leaseID }
}

func WithRevision(rev int64) KVOp {
	return func(opts *KVOptions) { opts.Rev = rev }
}

func (o *KVOptions) Apply(opts []KVOp) {
	for _, opt := range opts {
		opt(o)
	}
}

func GetKVOptions(opts []KVOp) *KVOptions {
	o := KVOptions{}
	o.Apply(opts)
	return &o
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

type SyncCb func(ctx context.Context, data *SyncCbData)

type SyncCbData struct {
	// action on the data
	Action SyncCbAction
	// key - will be nil for action SyncListAllStart/End/Error
	Key []byte
	// value - will be nil for action SyncListAllStart/End/Error
	Value []byte
	// global revision of the data
	Rev int64
	// last modified revision of the data
	ModRev int64
	// MoreEvents indicates there are more changes in the revision.
	// With transactions, multiple changes can be done in the same
	// revision, but each change is called back separately.
	// MoreEvents is set to true if there are more changes to be
	// called back for the current revision.
	MoreEvents bool
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
	// Single element in db - type is the key
	if ii == -1 {
		return region, key, "", nil
	}
	typ = key[:ii]
	key = key[ii+1:]

	return region, typ, key, nil
}

func NotFoundError(key string) error {
	return fmt.Errorf("key %s not found", DbKeyPrefixRemove(key))
}

func ExistsError(key string) error {
	return fmt.Errorf("key %s already exists", DbKeyPrefixRemove(key))
}
