package cache

type ICache interface {
	Store(key string, data string) error
	Load(key string) (string, error)
	IsCacheDisabled() bool
}

func New(noCache bool) ICache {
	return &FileBasedCache{
		noCache: noCache,
	}
}
