package cache

import (
	rpc "buf.build/gen/go/interplex-ai/schemas/grpc/go/protobuf/schema/v1/schemav1grpc"
	schemav1 "buf.build/gen/go/interplex-ai/schemas/protocolbuffers/go/protobuf/schema/v1"
	"context"
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
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
	conn *grpc.ClientConn
}

func (c *InterplexClient) Close() error {
	return c.conn.Close()
}

func NewClient(address string) (*InterplexClient, error) {
	// Connect to the K8sGPT server and create a new client
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("failed to create context: %v", err)
	}
	// Wait until the connection is ready
	state := conn.GetState()
	if state != connectivity.Ready {
		return nil, fmt.Errorf("failed to connect, current state: %v", state)
	}
	client := &InterplexClient{conn: conn}

	return client, nil
}

func (c *InterplexCache) Configure(cacheInfo CacheProvider) error {

	if cacheInfo.Interplex.ConnectionString == "" {
		return errors.New("connection string is required")
	}
	c.configuration.ConnectionString = cacheInfo.Interplex.ConnectionString
	return nil
}

func (c *InterplexCache) Store(key string, data string) error {

	client, err := NewClient(c.configuration.ConnectionString)
	if err != nil {
		return err
	}
	c.client = *client
	serviceClient := rpc.NewCacheServiceClient(c.client.conn)
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
	client, err := NewClient(c.configuration.ConnectionString)
	if err != nil {
		return "", err
	}
	c.client = *client
	serviceClient := rpc.NewCacheServiceClient(c.client.conn)
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
