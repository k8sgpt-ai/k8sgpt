package googlecloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/spf13/viper"
)

const (
	securityBulletinEventURL = "type.googleapis.com/google.container.v1beta1.SecurityBulletinEvent"
	upgradeAvailableEventURL = "type.googleapis.com/google.container.v1beta1.UpgradeAvailableEvent"
	upgradeEventURL          = "type.googleapis.com/google.container.v1beta1.UpgradeEvent"
)

type GKEAnalyzer struct {
	EnableClusterNotificationAnalysis bool `mapstructure:"enable_cluster_notification_analysis"`
}

type GKEClusterNotificationAnalysis struct {
	ProjectID                         string `mapstructure:"project_id"`
	ClusterNotificationSubscriptionID string `mapstructure:"cluster_notification_subscription_id"`
}

type clusterNotification struct {
	ProjectID       string                 `json:"project_id"`
	ClusterLocation string                 `json:"cluster_location"`
	ClusterName     string                 `json:"cluster_name"`
	TypeURL         string                 `json:"type_url"`
	Payload         map[string]interface{} `json:"payload"`
}

func (g *GKEAnalyzer) Analyze(analysis common.Analyzer) ([]common.Result, error) {
	var cr []common.Result
	if g.EnableClusterNotificationAnalysis {
		var gkeClusterNotificationAnalysisList []GKEClusterNotificationAnalysis
		if err := viper.UnmarshalKey("googlecloud.gke.cluster_notification_analysis_list", &gkeClusterNotificationAnalysisList); err != nil {
			return nil, err
		}
		for _, gkeClusterNotificationAnalysis := range gkeClusterNotificationAnalysisList {
			if gkeClusterNotificationAnalysis.ProjectID == "" || gkeClusterNotificationAnalysis.ClusterNotificationSubscriptionID == "" {
				continue
			}
			pubsubClient, err := newPubSubClient(analysis.Context, gkeClusterNotificationAnalysis.ProjectID, gkeClusterNotificationAnalysis.ClusterNotificationSubscriptionID)
			if err != nil {
				return nil, err
			}
			pullSubscriptionResp, err := pubsubClient.PullSubscription(analysis.Context)
			if err != nil {
				return nil, err
			}
			if pullSubscriptionResp == nil {
				continue
			}
			for _, msg := range pullSubscriptionResp.ReceivedMessages {
				result, err := g.analyzeClusterNotification(&msg)
				if err != nil {
					return nil, err
				}
				cr = append(cr, result...)
			}
		}
	}
	analysis.Results = append(analysis.Results, cr...)
	return cr, nil
}

func (g *GKEAnalyzer) analyzeClusterNotification(msg *PubSubReceivedMessage) ([]common.Result, error) {
	var cr []common.Result
	// Unmarshal the PubSubReceivedMessage (ClusterNotification)
	var payload map[string]interface{}
	err := json.Unmarshal([]byte(msg.Message.Attributes["payload"]), &payload)
	if err != nil {
		return nil, err
	}
	clusterNotification := clusterNotification{
		ProjectID:       msg.Message.Attributes["project_id"],
		ClusterLocation: msg.Message.Attributes["cluster_location"],
		ClusterName:     msg.Message.Attributes["cluster_name"],
		TypeURL:         msg.Message.Attributes["type_url"],
		Payload:         payload,
	}
	// Indent the payload for better readability
	var buf bytes.Buffer
	if err := json.Indent(&buf, []byte(msg.Message.Attributes["payload"]), "", "  "); err != nil {
		return nil, err
	}
	switch clusterNotification.TypeURL {
	case securityBulletinEventURL:
		cr = append(cr, common.Result{
			Kind: "GKEClusterNotificationSecurityBulletinEvent",
			Name: "SecurityBulletinEvent",
			Error: []common.Failure{
				{Text: fmt.Sprintf(
					"project_id: %s\ncluster_location: %s\ncluster_name: %s\npayload: %s\n",
					clusterNotification.ProjectID, clusterNotification.ClusterLocation, clusterNotification.ClusterName, buf.String(),
				), KubernetesDoc: "", Sensitive: nil},
			},
		})
	case upgradeAvailableEventURL:
		cr = append(cr, common.Result{
			Kind: "GKEClusterNotificationUpgradeAvailabilityEvent",
			Name: "UpgradeAvailableEvent",
			Error: []common.Failure{
				{Text: fmt.Sprintf(
					"project_id: %s\ncluster_location: %s\ncluster_name: %s\npayload: %s\n",
					clusterNotification.ProjectID, clusterNotification.ClusterLocation, clusterNotification.ClusterName, buf.String()),
					KubernetesDoc: "", Sensitive: nil},
			},
		})
	case upgradeEventURL:
		cr = append(cr, common.Result{
			Kind: "GKEClusterNotificationUpgradeEvent",
			Name: "UpgradeEvent",
			Error: []common.Failure{
				{Text: fmt.Sprintf(
					"project_id: %s\ncluster_location: %s\ncluster_name: %s\npayload: %s\n",
					clusterNotification.ProjectID, clusterNotification.ClusterLocation, clusterNotification.ClusterName, buf.String()),
					KubernetesDoc: "", Sensitive: nil},
			},
		})
	}
	return cr, nil
}
