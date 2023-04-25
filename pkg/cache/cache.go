package cache

type ICache interface {
	Store(key string, data string) error
	Load(key string) (string, error)
	Exists(key string) bool
}

func New(noCache bool) ICache {
	if noCache {
		return &NoopCache{}
	}

	return &FileBasedCache{}
}
