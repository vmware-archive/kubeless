package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"testing"

	v2beta1 "k8s.io/api/autoscaling/v2beta1"
	"k8s.io/api/extensions/v1beta1"
	extensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	fakeextensionsapi "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ktesting "k8s.io/client-go/testing"
)

func objBody(object interface{}) io.ReadCloser {
	output, err := json.Marshal(object)
	if err != nil {
		panic(err)
	}
	return ioutil.NopCloser(bytes.NewReader([]byte(output)))
}

func fakeConfig() *rest.Config {
	return &rest.Config{
		Host: "https://example.com:443",
		ContentConfig: rest.ContentConfig{
			GroupVersion: &schema.GroupVersion{
				Group:   "",
				Version: "v1",
			},
			NegotiatedSerializer: scheme.Codecs,
		},
	}
}

func TestGetLocalHostname(t *testing.T) {
	config := fakeConfig()
	expectedHostName := "foobar.example.com.nip.io"
	actualHostName, err := GetLocalHostname(config, "foobar")
	if err != nil {
		t.Error(err)
	}

	if expectedHostName != actualHostName {
		t.Errorf("Expected %s but got %s", expectedHostName, actualHostName)
	}
}

func TestCreateAutoscaleResource(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	name := "foo"
	ns := "myns"
	hpaDef := v2beta1.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
	}
	if err := CreateAutoscale(clientset, hpaDef); err != nil {
		t.Fatalf("Creating autoscale returned err: %v", err)
	}

	hpa, err := clientset.AutoscalingV2beta1().HorizontalPodAutoscalers(ns).Get(name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Creating autoscale returned err: %v", err)
	}
	if hpa.ObjectMeta.Name != "foo" {
		t.Fatalf("Creating wrong scale target name")
	}
}

func TestDeleteAutoscaleResource(t *testing.T) {
	myNsFoo := metav1.ObjectMeta{
		Namespace: "myns",
		Name:      "foo",
	}

	as := v2beta1.HorizontalPodAutoscaler{
		ObjectMeta: myNsFoo,
	}

	clientset := fake.NewSimpleClientset(&as)
	if err := DeleteAutoscale(clientset, "foo", "myns"); err != nil {
		t.Fatalf("Deleting autoscale returned err: %v", err)
	}
	a := clientset.Actions()
	if ns := a[0].GetNamespace(); ns != "myns" {
		t.Errorf("deleted autoscale from wrong namespace (%s)", ns)
	}
	if name := a[0].(ktesting.DeleteAction).GetName(); name != "foo" {
		t.Errorf("deleted autoscale with wrong name (%s)", name)
	}
}

func TestInitializeEmptyMapsInDeployment(t *testing.T) {
	deployment := v1beta1.Deployment{}
	deployment.Spec.Selector = &metav1.LabelSelector{}
	initializeEmptyMapsInDeployment(&deployment)
	if deployment.ObjectMeta.Annotations == nil {
		t.Fatal("ObjectMeta.Annotations map is nil")
	}
	if deployment.ObjectMeta.Labels == nil {
		t.Fatal("ObjectMeta.Labels map is nil")
	}
	if deployment.Spec.Selector == nil && deployment.Spec.Selector.MatchLabels == nil {
		t.Fatal("deployment.Spec.Selector.MatchLabels is nil")
	}
	if deployment.Spec.Template.ObjectMeta.Labels == nil {
		t.Fatal("deployment.Spec.Template.ObjectMeta.Labels map is nil")
	}
	if deployment.Spec.Template.ObjectMeta.Annotations == nil {
		t.Fatal("deployment.Spec.Template.ObjectMeta.Annotations map is nil")
	}
	if deployment.Spec.Template.Spec.NodeSelector == nil {
		t.Fatal("deployment.Spec.Template.Spec.NodeSelector map is nil")
	}
}

func TestMergeDeployments(t *testing.T) {
	var replicas int32
	replicas = 10
	destinationDeployment := v1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"foo1-deploy": "bar",
			},
		},
	}

	sourceDeployment := v1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"foo2-deploy": "bar",
			},
		},
		Spec: v1beta1.DeploymentSpec{
			Replicas: &replicas,
		},
	}

	MergeDeployments(&destinationDeployment, &sourceDeployment)
	expectedAnnotations := map[string]string{
		"foo1-deploy": "bar",
		"foo2-deploy": "bar",
	}
	for i := range expectedAnnotations {
		if destinationDeployment.ObjectMeta.Annotations[i] != expectedAnnotations[i] {
			t.Fatalf("Expecting annotation %s but received %s", destinationDeployment.ObjectMeta.Annotations[i], expectedAnnotations[i])
		}
	}
	if *destinationDeployment.Spec.Replicas != replicas {
		t.Fatalf("Expecting replicas as 10 but received %v", *destinationDeployment.Spec.Replicas)
	}

}

func TestGetAnnotationsFromCRD(t *testing.T) {
	crdWithoutAnnotationName := "crdWithoutAnnotation"
	crdWithAnnotationName := "crdWithAnnotation"
	expectedAnnotations := map[string]string{
		"foo": "bar",
	}
	crdWithAnnotation := &extensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"foo": "bar",
			},
			Name: crdWithAnnotationName,
		},
		Spec: extensionsv1beta1.CustomResourceDefinitionSpec{
			Group: "foo.group.io",
			Names: extensionsv1beta1.CustomResourceDefinitionNames{
				Plural:   "foos",
				Singular: "foo",
				Kind:     "fooKind",
				ListKind: "fooList",
			},
		},
	}
	clientset := fakeextensionsapi.NewSimpleClientset()
	_, err := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crdWithAnnotation)
	if err != nil {
		t.Fatalf("Error while creating CRD: %v", err)
	}
	annotations, err := GetAnnotationsFromCRD(clientset, crdWithAnnotationName)
	if err != nil {
		t.Fatalf("Error while fetching CRD: %v", err)
	}
	for i := range expectedAnnotations {
		if annotations[i] != expectedAnnotations[i] {
			t.Errorf("Expecting annotation %s but received %s", expectedAnnotations[i], annotations[i])
		}
	}

	crdWithoutAnnotation := &extensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{},
			Name:        crdWithoutAnnotationName,
		},
		Spec: extensionsv1beta1.CustomResourceDefinitionSpec{
			Group: "foo.group.io",
			Names: extensionsv1beta1.CustomResourceDefinitionNames{
				Plural:   "foos",
				Singular: "foo",
				Kind:     "fooKind",
				ListKind: "fooList",
			},
		},
	}
	_, err = clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crdWithoutAnnotation)
	if err != nil {
		t.Fatalf("Error while creating CRD: %v", err)
	}
	annotations, err = GetAnnotationsFromCRD(clientset, crdWithoutAnnotationName)
	if err != nil {
		t.Fatalf("Error while fetching annotations from CRD: %v", err)
	}
	if len(annotations) != 0 {
		t.Errorf("Expecting annotations of length 0 but received length %d", len(annotations))
	}

}
