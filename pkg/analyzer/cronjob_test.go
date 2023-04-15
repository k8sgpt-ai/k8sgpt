package analyzer

import (
	"context"
	"testing"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/magiconair/properties/assert"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestCronJobSuccess(t *testing.T) {
	clientset := fake.NewSimpleClientset(&batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-cronjob",
			Namespace: "default",
			Annotations: map[string]string{
				"analysisDate": "2022-04-01",
			},
			Labels: map[string]string{
				"app": "example-app",
			},
		},
		Spec: batchv1.CronJobSpec{
			Schedule:          "*/1 * * * *",
			ConcurrencyPolicy: "Allow",
			JobTemplate: batchv1.JobTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "example-app",
					},
				},
				Spec: batchv1.JobSpec{
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Name:  "example-container",
									Image: "nginx",
								},
							},
							RestartPolicy: v1.RestartPolicyOnFailure,
						},
					},
				},
			},
		},
	})

	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: clientset,
		},
		Context:   context.Background(),
		Namespace: "default",
	}

	analyzer := CronJobAnalyzer{}
	analysisResults, err := analyzer.Analyze(config)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, len(analysisResults), 0)
}

func TestCronJobBroken(t *testing.T) {
	clientset := fake.NewSimpleClientset(&batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-cronjob",
			Namespace: "default",
			Annotations: map[string]string{
				"analysisDate": "2022-04-01",
			},
			Labels: map[string]string{
				"app": "example-app",
			},
		},
		Spec: batchv1.CronJobSpec{
			Schedule:          "*** * * * *",
			ConcurrencyPolicy: "Allow",
			JobTemplate: batchv1.JobTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "example-app",
					},
				},
				Spec: batchv1.JobSpec{
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Name:  "example-container",
									Image: "nginx",
								},
							},
							RestartPolicy: v1.RestartPolicyOnFailure,
						},
					},
				},
			},
		},
	})

	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: clientset,
		},
		Context:   context.Background(),
		Namespace: "default",
	}

	analyzer := CronJobAnalyzer{}
	analysisResults, err := analyzer.Analyze(config)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, len(analysisResults), 1)
	assert.Equal(t, analysisResults[0].Name, "default/example-cronjob")
	assert.Equal(t, analysisResults[0].Kind, "CronJob")
}
