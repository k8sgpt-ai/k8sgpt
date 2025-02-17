package custom

import (
	rpc "github.com/naveenthangaraj03/k8sgpt-custom/proto/schema1/v1"
	"context"
	"fmt"
	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	c              *grpc.ClientConn
	analyzerClient rpc.CustomAnalyzerServiceClient
}

func NewClient(c Connection) (*Client, error) {

	//nolint:staticcheck // Ignoring SA1019 for compatibility reasons
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", c.Url, c.Port), grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, err
	}
	client := rpc.NewCustomAnalyzerServiceClient(conn)
	return &Client{
		c:              conn,
		analyzerClient: client,
	}, nil
}

func (cli *Client) Run(filter string) (common.Result, error) {
	var result common.Result
	req := &rpc.RunRequest{
		Filter: filter,
	}
	res, err := cli.analyzerClient.Run(context.Background(), req)
	if err != nil {
		return result, err
	}
	if res.Result != nil {

		// We should refactor this, because Error and Failure do not map 1:1 from K8sGPT/schema
		var errorsFound []common.Failure
		for _, e := range res.Result.Error {
			errorsFound = append(errorsFound, common.Failure{
				Text: e.Text,
				// TODO: Support sensitive data
			})
		}

		result.Name = res.Result.Name
		result.Kind = res.Result.Kind
		result.Details = res.Result.Details
		result.ParentObject = res.Result.ParentObject
		result.Error = errorsFound
	}
	return result, nil
}
