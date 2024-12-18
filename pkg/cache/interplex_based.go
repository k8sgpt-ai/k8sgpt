package cache

import (
	rpc "buf.build/gen/go/interplex-ai/schemas/grpc/go/protobuf/schema/v1/schemav1grpc"
	schemav1 "buf.build/gen/go/interplex-ai/schemas/protocolbuffers/go/protobuf/schema/v1"
	"context"
	"errors"
	"google.golang.org/grpc"
	"os"
)

var _ ICache = (*InterplexCache)(nil)

type InterplexCache struct {
	configuration      InterplexCacheConfiguration
	client             InterplexClient
	cacheServiceClient rpc.CacheServiceClient
	noCache            bool
}

type InterplexCacheConfiguration struct {
	ConnectionString string `mapstructure:"connectionString" yaml:"connectionString,omitempty"`
}

type InterplexClient struct {
}

func (c *InterplexCache) Configure(cacheInfo CacheProvider) error {

	if cacheInfo.Interplex.ConnectionString == "" {
		return errors.New("connection string is required")
	}
	c.configuration.ConnectionString = cacheInfo.Interplex.ConnectionString
	return nil
}

func (c *InterplexCache) Store(key string, data string) error {

	if os.Getenv("INTERPLEX_LOCAL_MODE") != "" {
		c.configuration.ConnectionString = "localhost:8084"
	}

	conn, err := grpc.NewClient(c.configuration.ConnectionString, grpc.WithInsecure(), grpc.WithBlock())
	defer conn.Close()
	if err != nil {
		return err
	}
	serviceClient := rpc.NewCacheServiceClient(conn)
	c.cacheServiceClient = serviceClient
	req := schemav1.SetRequest{
		Key:   key,
		Value: data,
	}
	_, err = c.cacheServiceClient.Set(context.Background(), &req)
	if err != nil {
		return err
	}
	return nil
}

func (c *InterplexCache) Load(key string) (string, error) {
	conn, err := grpc.NewClient(c.configuration.ConnectionString, grpc.WithInsecure(), grpc.WithBlock())
	defer conn.Close()
	if err != nil {
		return "", err
	}
	serviceClient := rpc.NewCacheServiceClient(conn)
	c.cacheServiceClient = serviceClient
	req := schemav1.GetRequest{
		Key: key,
	}
	resp, err := c.cacheServiceClient.Get(context.Background(), &req)
	// check if response is cache error not found
	if err != nil {
		return "", err
	}
	return resp.Value, nil
}

func (InterplexCache) List() ([]CacheObjectDetails, error) {
	//TODO implement me
	return nil, errors.New("not implemented")
}

func (InterplexCache) Remove(key string) error {

	return errors.New("not implemented")
}

func (c *InterplexCache) Exists(key string) bool {
	if _, err := c.Load(key); err != nil {
		return false
	}
	return true
}

func (c *InterplexCache) IsCacheDisabled() bool {
	return c.noCache
}

func (InterplexCache) GetName() string {
	//TODO implement me
	return "interplex"
}

func (c *InterplexCache) DisableCache() {
	c.noCache = true
}
