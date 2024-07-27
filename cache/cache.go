package cache

type Cache interface {
	Set(key string, value interface{}) error
	Get(key string) (interface{}, error)
	Del(key string) error
	Exists(key string) (bool, error)
	Map() (map[string]interface{}, error)
	JSON() ([]byte, error)
	Debug(identifier string) error
	Flush() error
	DeleteByPrefix(prefix string) error
}
