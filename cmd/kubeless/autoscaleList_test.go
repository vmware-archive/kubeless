package main

import (
	"bytes"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/pkg/api/v1"
	av2alpha1 "k8s.io/client-go/pkg/apis/autoscaling/v2alpha1"
)

func listAutoscaleOutput(t *testing.T, client kubernetes.Interface, ns, output string) string {
	var buf bytes.Buffer

	if err := doAutoscaleList(&buf, client, ns, output); err != nil {
		t.Fatalf("doList returned error: %v", err)
	}

	return buf.String()
}

func TestAutoscaleList(t *testing.T) {
	replicas := int32(1)
	targetAverageUtilization := int32(50)
	as1 := av2alpha1.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "myns",
		},
		Spec: av2alpha1.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: av2alpha1.CrossVersionObjectReference{
				Kind: "Deployment",
				Name: "foo",
			},
			MinReplicas: &replicas,
			MaxReplicas: replicas,
			Metrics: []av2alpha1.MetricSpec{
				{
					Type: "Resource",
					Resource: &av2alpha1.ResourceMetricSource{
						Name: v1.ResourceCPU,
						TargetAverageUtilization: &targetAverageUtilization,
					},
				},
			},
		},
	}

	as2 := av2alpha1.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bar",
			Namespace: "myns",
		},
		Spec: av2alpha1.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: av2alpha1.CrossVersionObjectReference{
				Kind: "Deployment",
				Name: "foo",
			},
			MinReplicas: &replicas,
			MaxReplicas: replicas,
			Metrics: []av2alpha1.MetricSpec{
				{
					Type: "Resource",
					Resource: &av2alpha1.ResourceMetricSource{
						Name: v1.ResourceCPU,
						TargetAverageUtilization: &targetAverageUtilization,
					},
				},
			},
		},
	}

	client := fake.NewSimpleClientset(&as1, &as2)

	output := listAutoscaleOutput(t, client, "myns", "")
	t.Log("output is", output)

	if !strings.Contains(output, "foo") || !strings.Contains(output, "bar") {
		t.Errorf("table output didn't mention both functions")
	}

	// json output
	output = listAutoscaleOutput(t, client, "myns", "json")
	t.Log("output is", output)

	// yaml output
	output = listAutoscaleOutput(t, client, "myns", "yaml")
	t.Log("output is", output)
}
