package cache

type ICache interface {
	Store(key string, data string) error
	Load(key string) (string, error)
	Exists(key string) bool
	IsCacheDisabled() bool
}

func New(noCache bool) ICache {
	return &FileBasedCache{
		noCache: noCache,
	}
}
