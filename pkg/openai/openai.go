package openai

import (
	"fmt"

	"github.com/sashabaranov/go-openai"
	"github.com/spf13/viper"
)

type Client struct {
	client *openai.Client
}

func (c *Client) GetClient() *openai.Client {
	return c.client
}
func NewClient() (*Client, error) {

	// get the token with viper
	token := viper.GetString("openai_api_key")
	// check if nil
	if token == "" {
		return nil, fmt.Errorf("no OpenAI API Key found")
	}

	client := openai.NewClient(token)
	return &Client{
		client: client,
	}, nil
}
