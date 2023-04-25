package cache

var _ (ICache) = (*NoopCache)(nil)

type NoopCache struct{}

func (c *NoopCache) Store(key string, data string) error {
	return nil
}

func (c *NoopCache) Load(key string) (string, error) {
	return "", nil
}

func (c *NoopCache) Exists(key string) bool {
	return false
}
