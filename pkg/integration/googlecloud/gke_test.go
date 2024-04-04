package googlecloud

import (
	"fmt"
	"testing"
	"time"
)

func Test_GKEAnalyzer_analyzeClusterNotification(t *testing.T) {
	type args struct {
		msg PubSubReceivedMessage
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "Test_GKEAnalyzer_analyzeClusterNotification_SecurityBulletinEvent",
			args: args{
				msg: PubSubReceivedMessage{
					Message: struct {
						Data        []byte            `json:"data"`
						Attributes  map[string]string `json:"attributes"`
						MessageID   string            `json:"messageId"`
						PublishTime time.Time         `json:"publishTime"`
						OrderingKey string            `json:"orderingKey"`
					}(struct {
						Data        []byte
						Attributes  map[string]string
						MessageID   string
						PublishTime time.Time
						OrderingKey string
					}{
						Attributes: map[string]string{
							"project_id":       "123456789",
							"cluster_location": "us-central1-c",
							"cluster_name":     "example-cluster",
							"type_url":         "type.googleapis.com/google.container.v1beta1.SecurityBulletinEvent",
							"payload":          "{\"resourceTypeAffected\":\"RESOURCE_TYPE_CONTROLPLANE\",\"bulletinId\":\"GCP-2021-001\",\"cveIds\":[\"CVE-2021-3156\"],\"severity\":\"Medium\",\"briefDescription\":\"A vulnerability was recently discovered in the Linux utility sudo, described in CVE-2021-3156, that may allow an attacker with unprivileged local shell access on a system with sudo installed to escalate their privileges to root on the system.\",\"affectedSupportedMinors\":[\"1.18\",\"1.19\"],\"patchedVersions\":[\"1.18.9-gke.1900\",\"1.19.9-gke.1900\"],\"suggestedUpgradeTarget\":\"1.19.9-gke.1900\",\"bulletinUri\":\"https://cloud.google.com/anthos/clusters/docs/security-bulletins#gcp-2021-001\"}",
						},
					})},
			},
			want: 1,
		},
		{
			name: "Test_GKEAnalyzer_analyzeClusterNotification_UpgradeAvailableEvent",
			args: args{
				msg: PubSubReceivedMessage{Message: struct {
					Data        []byte            `json:"data"`
					Attributes  map[string]string `json:"attributes"`
					MessageID   string            `json:"messageId"`
					PublishTime time.Time         `json:"publishTime"`
					OrderingKey string            `json:"orderingKey"`
				}(struct {
					Data        []byte
					Attributes  map[string]string
					MessageID   string
					PublishTime time.Time
					OrderingKey string
				}{
					Attributes: map[string]string{
						"project_id":       "123456789",
						"cluster_location": "us-central1-c",
						"cluster_name":     "example-cluster",
						"type_url":         "type.googleapis.com/google.container.v1beta1.UpgradeAvailableEvent",
						"payload":          "{ \"version\":\"1.17.15-gke.800\",\n  \"resourceType\":\"MASTER\",\n  \"releaseChannel\":{\"channel\":\"RAPID\"},\n  \"windowsVersions\": [\n    {\n      \"imageType\": \"WINDOWS_SAC\",\n      \"osVersion\": \"10.0.18363.1198\",\n      \"supportEndDate\": {\n        \"day\": 10,\n        \"month\": 5,\n        \"year\": 2022\n      }\n    },\n    {\n      \"imageType\": \"WINDOWS_LTSC\",\n      \"osVersion\": \"10.0.17763.1577\",\n      \"supportEndDate\": {\n        \"day\": 9,\n        \"month\": 1,\n        \"year\": 2024\n      }\n    }\n  ]\n}",
					},
				})},
			},
			want: 1,
		},
		{
			name: "Test_GKEAnalyzer_analyzeClusterNotification_UpgradeEvent",
			args: args{
				msg: PubSubReceivedMessage{Message: struct {
					Data        []byte            `json:"data"`
					Attributes  map[string]string `json:"attributes"`
					MessageID   string            `json:"messageId"`
					PublishTime time.Time         `json:"publishTime"`
					OrderingKey string            `json:"orderingKey"`
				}(struct {
					Data        []byte
					Attributes  map[string]string
					MessageID   string
					PublishTime time.Time
					OrderingKey string
				}{
					Attributes: map[string]string{
						"project_id":       "123456789",
						"cluster_location": "us-central1-c",
						"cluster_name":     "example-cluster",
						"type_url":         "type.googleapis.com/google.container.v1beta1.UpgradeEvent",
						"payload":          "{\"resourceType\":\"MASTER\",\"operation\":\"operation-1595889094437-87b7254a\",\"operationStartTime\":\"2020-07-27T22:31:34.437652293Z\",\"currentVersion\":\"1.15.12-gke.2\",\"targetVersion\":\"1.15.12-gke.9\"}",
					},
				})},
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &GKEAnalyzer{
				EnableClusterNotificationAnalysis: true,
			}
			got, err := g.analyzeClusterNotification(&tt.args.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("analyzeClusterNotification() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for _, result := range got {
				fmt.Println(result)
			}
			if len(got) != tt.want {
				t.Errorf("analyzeClusterNotification() got = %v, want %v", len(got), tt.want)
			}
		})
	}
}
