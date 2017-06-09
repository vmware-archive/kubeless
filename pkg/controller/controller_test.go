package controller

import (
	"testing"

	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/pkg/api/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	xv1beta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
	ktesting "k8s.io/client-go/testing"
)

func findAction(fake *fake.Clientset, verb, resource string) ktesting.Action {
	for _, a := range fake.Actions() {
		if a.Matches(verb, resource) {
			return a
		}
	}
	return nil
}

func hasAction(fake *fake.Clientset, verb, resource string) bool {
	return findAction(fake, verb, resource) != nil
}

func TestInitResource(t *testing.T) {
	client := fake.NewSimpleClientset()
	if err := initResource(client); err != nil {
		t.Errorf("initResource() failed with %v", err)
	}

	if !hasAction(client, "create", "thirdpartyresources") {
		t.Errorf("initResource() failed to create thirdpartyresource")
	}

	client.ClearActions()

	if err := initResource(client); err != nil {
		t.Errorf("initResource() with existing TPR failed with %v", err)
	}

	if !hasAction(client, "update", "thirdpartyresources") {
		t.Errorf("initResource() failed to update existing thirdpartyresource")
	}
}

func TestCollectObjects(t *testing.T) {
	var (
		nonExistFunc types.UID = "00000000-0000-0000-0000-000000000000"
		bt                     = true
	)

	myNsFoo := metav1.ObjectMeta{
		Namespace: "myns",
		Name:      "foo",
		OwnerReferences: []metav1.OwnerReference{
			{
				Kind:               "Function",
				APIVersion:         "k8s.io",
				Name:               "funcFoo",
				UID:                nonExistFunc,
				BlockOwnerDeletion: &bt,
			},
		},
	}

	deploy := xv1beta1.Deployment{
		ObjectMeta: myNsFoo,
	}

	svc := v1.Service{
		ObjectMeta: myNsFoo,
	}

	cm := v1.ConfigMap{
		ObjectMeta: myNsFoo,
	}

	client := fake.NewSimpleClientset(&deploy, &svc, &cm)

	functionUIDSet := make(map[types.UID]bool)
	//functionUIDSet[nonExistFunc] = false

	if err := collectServices(client, functionUIDSet); err != nil {
		t.Errorf("collectService() failed with %v", err)
	}
	if err := collectDeployment(client, functionUIDSet); err != nil {
		t.Errorf("collectDeployment() failed with %v", err)
	}
	if err := collectConfigMap(client, functionUIDSet); err != nil {
		t.Errorf("collectConfigMap() failed with %v", err)
	}
}
