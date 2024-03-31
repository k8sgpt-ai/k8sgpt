package googlecloud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"golang.org/x/oauth2/google"
	"net/http"
	"os"
	"time"
)

/*
PubSubClient is a client for Google Cloud Pub/Sub.
PubSub Client is used to receive GKE cluster notifications from Cloud Pub/Sub;
the reason for using the REST endpoint instead of gRPC is that we wanted to query only as much as needed, without using gRPC Streaming.
*/

type PubSubClient struct {
	ProjectID      string
	SubscriptionID string
	HTTPClient     *http.Client
	TimeoutSec     int  `json:"timeout_sec"`
	EnableAck      bool `json:"enable_ack"`
	MaxMessages    int  `json:"max_messages"`
}

func newPubSubClient(ctx context.Context, projectID, subscriptionID string) (*PubSubClient, error) {
	googleDefaultClient, err := google.DefaultClient(ctx, []string{
		"https://www.googleapis.com/auth/pubsub",
		"https://www.googleapis.com/auth/cloud-platform"}...,
	)
	if err != nil {
		return nil, err
	}
	timeoutSec := 10
	if viper.Get("googlecloud.pubsub.timeout_sec") != nil {
		timeoutSec = viper.GetInt("googlecloud.pubsub.timeout_sec")
	}
	enableAcK := true
	if viper.Get("googlecloud.pubsub.enable_ack") != nil {
		enableAcK = viper.GetBool("googlecloud.pubsub.enable_ack")
	}
	maxMessage := 1
	if viper.Get("googlecloud.pubsub.max_messages") != nil {
		maxMessage = viper.GetInt("googlecloud.pubsub.max_messages")
	}
	return &PubSubClient{
		ProjectID:      projectID,
		SubscriptionID: subscriptionID,
		HTTPClient:     googleDefaultClient,
		TimeoutSec:     timeoutSec,
		EnableAck:      enableAcK,
		MaxMessages:    maxMessage,
	}, nil
}

type PullSubscriptionRequest struct {
	ReturnImmediately bool `json:"returnImmediately"`
	MaxMessages       int  `json:"maxMessages"`
}

type PullSubscriptionResponse struct {
	ReceivedMessages []PubSubReceivedMessage `json:"receivedMessages"`
}

type PubSubReceivedMessage struct {
	AckID   string `json:"ackId"`
	Message struct {
		Data        []byte            `json:"data"`
		Attributes  map[string]string `json:"attributes"`
		MessageID   string            `json:"messageId"`
		PublishTime time.Time         `json:"publishTime"`
		OrderingKey string            `json:"orderingKey"`
	} `json:"message"`
	DeliverAttempt int `json:"deliveryAttempt"`
}

func (p *PubSubClient) GetPullSubscriptionEndpoint() string {
	// PullSubscription Endpoint Example: https://pubsub.googleapis.com/v1/projects/myproject/subscriptions/mysubscription:pull
	return fmt.Sprintf("https://pubsub.googleapis.com/v1/projects/%s/subscriptions/%s:pull", p.ProjectID, p.SubscriptionID)
}

func (p *PubSubClient) PullSubscription(ctx context.Context) (*PullSubscriptionResponse, error) {
	reqBody := PullSubscriptionRequest{
		MaxMessages: p.MaxMessages,
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(reqBody); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(p.TimeoutSec)*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.GetPullSubscriptionEndpoint(), &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		if os.IsTimeout(err) {
			fmt.Println("error by timeout. maybe message is not found.")
			return nil, nil
		}
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("pull subscription failed: %s", resp.Status)
	}
	defer resp.Body.Close()
	var pullSubscriptionResponse PullSubscriptionResponse
	if err := json.NewDecoder(resp.Body).Decode(&pullSubscriptionResponse); err != nil {
		return nil, err
	}
	if p.EnableAck {
		ackIDs := make([]string, 0, len(pullSubscriptionResponse.ReceivedMessages))
		for _, msg := range pullSubscriptionResponse.ReceivedMessages {
			ackIDs = append(ackIDs, msg.AckID)
		}
		if err := p.AckSubscription(ctx, ackIDs); err != nil {
			return nil, err
		}
	}
	return &pullSubscriptionResponse, nil
}

type AckSubscriptionRequest struct {
	AckIDs []string `json:"ackIds"`
}

func (p *PubSubClient) GetAckSubscriptionEndpoint() string {
	// AckSubscription Endpoint Example: https://pubsub.googleapis.com/v1/projects/myproject/subscriptions/mysubscription:acknowledge
	return fmt.Sprintf("https://pubsub.googleapis.com/v1/projects/%s/subscriptions/%s:acknowledge", p.ProjectID, p.SubscriptionID)
}

func (p *PubSubClient) AckSubscription(ctx context.Context, ackIDs []string) error {
	reqBody := AckSubscriptionRequest{
		AckIDs: ackIDs,
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(reqBody); err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.GetAckSubscriptionEndpoint(), &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("acknowledge subscription failed: %s", resp.Status)
	}
	return nil
}
