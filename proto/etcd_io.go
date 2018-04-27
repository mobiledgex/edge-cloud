package proto

type EtcdIO interface {
	Create(key, val string) error
	Update(key, val string, version int64) error
	Delete(key string) error
	Get(key string) ([]byte, int64, error)
}
