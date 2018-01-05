package autoscale

import (
	"bytes"
	"strings"
	"testing"

	av2alpha1 "k8s.io/api/autoscaling/v2beta1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
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
	q, _ := resource.ParseQuantity("10k")

	as1 := av2alpha1.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "myns",
			Labels: map[string]string{
				"created-by": "kubeless",
			},
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
					Type: av2alpha1.ResourceMetricSourceType,
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
			Labels: map[string]string{
				"created-by": "kubeless",
			},
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
					Type: av2alpha1.ObjectMetricSourceType,
					Object: &av2alpha1.ObjectMetricSource{
						MetricName:  "function_calls",
						TargetValue: q,
						Target: av2alpha1.CrossVersionObjectReference{
							Kind: "Service",
							Name: "foo",
						},
					},
				},
			},
		},
	}

	as3 := av2alpha1.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foobar",
			Namespace: "myns",
		},
	}

	client := fake.NewSimpleClientset(&as1, &as2, &as3)

	output := listAutoscaleOutput(t, client, "myns", "")
	t.Log("output is", output)

	if !strings.Contains(output, "foo") || !strings.Contains(output, "bar") {
		t.Errorf("table output didn't mention both autoscales")
	}

	if strings.Contains(output, "foobar") {
		t.Errorf("table output shouldn't mention foobar autoscale as it isn't created by kubeless")
	}

	// json output
	output = listAutoscaleOutput(t, client, "myns", "json")
	t.Log("output is", output)

	// yaml output
	output = listAutoscaleOutput(t, client, "myns", "yaml")
	t.Log("output is", output)
}
