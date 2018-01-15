package autoscale

import (
	"reflect"
	"testing"

	"k8s.io/api/autoscaling/v2beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetHorizontalAutoscaleDefinition(t *testing.T) {
	var min, max int32
	min = 1
	max = 3
	funcName := "test-autoscale"
	ns := "default"
	value := "10"
	labels := map[string]string{
		"foo": "bar",
	}
	metric := "cpu"
	hpa, err := getHorizontalAutoscaleDefinition(funcName, ns, metric, min, max, value, labels)
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	expectedMeta := metav1.ObjectMeta{
		Name:      funcName,
		Namespace: ns,
		Labels:    labels,
	}
	if hpa.Spec.ScaleTargetRef.Name != funcName {
		t.Fatalf("Creating wrong scale target name")
	}
	if !reflect.DeepEqual(expectedMeta, hpa.ObjectMeta) {
		t.Errorf("Expected \n%v to be equal to \n%v", expectedMeta, hpa.ObjectMeta)
	}
	if *hpa.Spec.MinReplicas != min {
		t.Errorf("Unexpected min replicas. Expecting %d got %d", min, *hpa.Spec.MinReplicas)
	}
	if hpa.Spec.MaxReplicas != max {
		t.Errorf("Unexpected max replicas. Expecting %d got %d", max, hpa.Spec.MaxReplicas)
	}
	if hpa.Spec.Metrics[0].Type != v2beta1.ResourceMetricSourceType ||
		*hpa.Spec.Metrics[0].Resource.TargetAverageUtilization != int32(10) {
		t.Error("Unexpected metric")
	}

	metric = "qps"
	hpa, err = getHorizontalAutoscaleDefinition(funcName, ns, metric, min, max, value, labels)
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	if hpa.Spec.Metrics[0].Type != v2beta1.ObjectMetricSourceType ||
		hpa.Spec.Metrics[0].Object.TargetValue.String() != "10" {
		t.Error("Unexpected metric")
	}
}
