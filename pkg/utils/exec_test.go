package utils

import (
	"testing"

	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func TestExecURL(t *testing.T) {
	conf := rest.Config{
		Host: "https://example.com/",
	}
	clientset := kubernetes.NewForConfigOrDie(&conf)

	opts := v1.PodExecOptions{
		Container: "ctr",
		Stderr:    true,
		Command:   []string{"a", "b"},
	}
	req, err := Exec(clientset.Core(), "mypod", "myns", opts)
	if err != nil {
		t.Fatal("Exec error:", err)
	}
	t.Logf("Got URL %v", req.URL)
	if req.URL.String() != "wss://example.com/api/v1/namespaces/myns/pods/mypod/exec?command=a&command=b&container=ctr&stderr=true" {
		t.Error("Unexpected url:", req.URL)
	}
}
