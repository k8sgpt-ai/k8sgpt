package cache

import (
	"context"
	"errors"
	"os"

	rpc "buf.build/gen/go/interplex-ai/schemas/grpc/go/protobuf/schema/v1/schemav1grpc"
	schemav1 "buf.build/gen/go/interplex-ai/schemas/protocolbuffers/go/protobuf/schema/v1"
	"google.golang.org/grpc"
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
	if os.Getenv("INTERPLEX_LOCAL_MODE") != "" {
		c.configuration.ConnectionString = "localhost:8084"
	}

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
	if err != nil {
		return "", err
	}
	return resp.Value, nil
}

func (c *InterplexCache) List() ([]CacheObjectDetails, error) {
	// Not implemented for Interplex cache
	return []CacheObjectDetails{}, nil
}

func (c *InterplexCache) Remove(key string) error {
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
	req := schemav1.DeleteRequest{
		Key: key,
	}
	_, err = c.cacheServiceClient.Delete(context.Background(), &req)
	return err
}

func (c *InterplexCache) Exists(key string) bool {
	_, err := c.Load(key)
	return err == nil
}

func (c *InterplexCache) IsCacheDisabled() bool {
	return c.noCache
}

func (c *InterplexCache) GetName() string {
	return "interplex"
}

func (c *InterplexCache) DisableCache() {
	c.noCache = true
}
