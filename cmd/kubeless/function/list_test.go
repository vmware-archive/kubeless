/*
Copyright (c) 2016-2017 Bitnami

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package function

import (
	"bytes"
	"regexp"
	"strings"
	"testing"

	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"

	kubelessApi "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	"github.com/kubeless/kubeless/pkg/client/clientset/versioned"
	fFake "github.com/kubeless/kubeless/pkg/client/clientset/versioned/fake"
)

func listOutput(t *testing.T, client versioned.Interface, apiV1Client kubernetes.Interface, ns, output string, args []string) string {
	var buf bytes.Buffer

	if err := doList(&buf, client, apiV1Client, ns, output, args); err != nil {
		t.Fatalf("doList returned error: %v", err)
	}

	return buf.String()
}

func TestList(t *testing.T) {
	funcMem, _ := parseResource("128Mi")
	listObj := kubelessApi.FunctionList{
		Items: []*kubelessApi.Function{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "myns",
				},
				Spec: kubelessApi.FunctionSpec{
					Handler:  "fhandler",
					Function: "ffunction",
					Runtime:  "fruntime",
					Deps:     "fdeps",
					Deployment: v1beta1.Deployment{
						Spec: v1beta1.DeploymentSpec{
							Template: v1.PodTemplateSpec{
								Spec: v1.PodSpec{
									Containers: []v1.Container{{}},
								},
							},
						},
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "bar",
					Namespace: "myns",
					Labels: map[string]string{
						"foo": "bar",
					},
				},
				Spec: kubelessApi.FunctionSpec{
					Handler:  "bhandler",
					Function: "bfunction",
					Runtime:  "nodejs6",
					Deps:     "{\"dependencies\": {\"test\": \"^1.0.0\"}}",
					Deployment: v1beta1.Deployment{
						Spec: v1beta1.DeploymentSpec{
							Template: v1.PodTemplateSpec{
								Spec: v1.PodSpec{
									Containers: []v1.Container{
										{
											Env: []v1.EnvVar{
												{
													Name:  "foo",
													Value: "bar",
												},
												{
													Name:  "foo2",
													Value: "bar2",
												},
											},
											Resources: v1.ResourceRequirements{
												Limits: map[v1.ResourceName]resource.Quantity{
													v1.ResourceMemory: funcMem,
												},
												Requests: map[v1.ResourceName]resource.Quantity{
													v1.ResourceMemory: funcMem,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "wrong",
					Namespace: "myns",
				},
				Spec: kubelessApi.FunctionSpec{
					Handler:  "fhandler",
					Function: "ffunction",
					Runtime:  "fruntime",
					Deps:     "fdeps",
					Deployment: v1beta1.Deployment{
						Spec: v1beta1.DeploymentSpec{
							Template: v1.PodTemplateSpec{
								Spec: v1.PodSpec{
									Containers: []v1.Container{{}},
								},
							},
						},
					},
				},
			},
		},
	}

	client := fFake.NewSimpleClientset(listObj.Items[0], listObj.Items[1], listObj.Items[2])

	deploymentFoo := v1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "myns",
		},
		Status: v1beta1.DeploymentStatus{
			Replicas:      int32(1),
			ReadyReplicas: int32(1),
		},
	}
	deploymentBar := v1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bar",
			Namespace: "myns",
		},
		Status: v1beta1.DeploymentStatus{
			Replicas:      int32(2),
			ReadyReplicas: int32(0),
		},
	}
	apiV1Client := fake.NewSimpleClientset(&deploymentFoo, &deploymentBar)

	// No arg -> list everything in namespace
	output := listOutput(t, client, apiV1Client, "myns", "", []string{})
	t.Log("output is", output)

	if !strings.Contains(output, "foo") || !strings.Contains(output, "bar") {
		t.Errorf("table output didn't mention both functions")
	}
	// Status
	m, err := regexp.MatchString("foo.*1/1 READY", output)
	if err != nil {
		t.Fatal(err)
	}
	if !m {
		t.Errorf("table output didn't mention deployment status")
	}
	m, err = regexp.MatchString("bar.*0/2 NOT READY", output)
	if err != nil {
		t.Fatal(err)
	}
	if !m {
		t.Errorf("table output didn't mention deployment status")
	}
	m, err = regexp.MatchString("wrong.*MISSING", output)
	if err != nil {
		t.Fatal(err)
	}
	if !m {
		t.Errorf("table output didn't mention deployment status")
	}

	// Explicit arg(s)
	output = listOutput(t, client, apiV1Client, "myns", "", []string{"foo"})
	t.Log("output is", output)

	if !strings.Contains(output, "foo") {
		t.Errorf("table output didn't mention explicit function foo")
	}
	if strings.Contains(output, "bar") {
		t.Errorf("table output mentions unrequested function bar")
	}

	if strings.Contains(output, "test: ^1.0.0") {
		t.Errorf("table output doesn't show parsed dependencies")
	}

	// TODO: Actually validate the output of the following.
	// Probably need to fix output framing first.

	// json output
	output = listOutput(t, client, apiV1Client, "myns", "json", []string{})
	t.Log("output is", output)
	if !strings.Contains(output, "foo") || !strings.Contains(output, "bar") {
		t.Errorf("table output didn't mention both functions")
	}

	// yaml output
	output = listOutput(t, client, apiV1Client, "myns", "yaml", []string{})
	t.Log("output is", output)
	if !strings.Contains(output, "128Mi") {
		t.Errorf("table output didn't mention proper memory of function")
	}

	// wide output
	output = listOutput(t, client, apiV1Client, "myns", "wide", []string{})
	t.Log("output is", output)
	if !strings.Contains(output, "foo = bar") {
		t.Errorf("table output didn't mention proper env of function")
	}
}
