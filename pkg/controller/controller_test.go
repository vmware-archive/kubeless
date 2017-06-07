package controller

import (
	"testing"

	"k8s.io/client-go/kubernetes/fake"
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
