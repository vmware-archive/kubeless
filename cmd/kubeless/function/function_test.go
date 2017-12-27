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
	"archive/zip"
	"io"
	"io/ioutil"
	"k8s.io/client-go/pkg/api/v1"
	"os"
	"reflect"
	"testing"

	"github.com/kubeless/kubeless/pkg/spec"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/fake"
)

func TestParseLabel(t *testing.T) {
	labels := []string{
		"foo=bar",
		"bar:foo",
		"foobar",
	}
	expected := map[string]string{
		"foo":    "bar",
		"bar":    "foo",
		"foobar": "",
	}
	actual := parseLabel(labels)
	if eq := reflect.DeepEqual(expected, actual); !eq {
		t.Errorf("Expect %v got %v", expected, actual)
	}
}

func TestParseEnv(t *testing.T) {
	envs := []string{
		"foo=bar",
		"bar:foo",
		"foobar",
		"foo=bar=baz",
		"qux=bar,baz",
	}
	expected := []v1.EnvVar{
		{
			Name:  "foo",
			Value: "bar",
		},
		{
			Name:  "bar",
			Value: "foo",
		},
		{
			Name:  "foobar",
			Value: "",
		},
		{
			Name:  "foo",
			Value: "bar=baz",
		},
		{
			Name:  "qux",
			Value: "bar,baz",
		},
	}
	actual := parseEnv(envs)
	if eq := reflect.DeepEqual(expected, actual); !eq {
		t.Errorf("Expect %v got %v", expected, actual)
	}
}

func TestGetFunctionDescription(t *testing.T) {
	// It should parse the given values
	file, err := ioutil.TempFile("", "test")
	if err != nil {
		t.Error(err)
	}
	_, err = file.WriteString("function")
	if err != nil {
		t.Error(err)
	}
	file.Close()
	defer os.Remove(file.Name()) // clean up

	inputHeadless := true
	inputPort := int32(80)
	result, err := getFunctionDescription(fake.NewSimpleClientset(), "test", "default", "file.handler", file.Name(), "dependencies", "runtime", "", "", "test-image", "128Mi", "10", true, &inputHeadless, &inputPort, []string{"TEST=1"}, []string{"test=1"}, spec.Function{})
	if err != nil {
		t.Error(err)
	}
	parsedMem, _ := parseMemory("128Mi")
	expectedFunction := spec.Function{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Function",
			APIVersion: "k8s.io/v1",
		},
		Metadata: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
			Labels: map[string]string{
				"test": "1",
			},
		},
		Spec: spec.FunctionSpec{
			Handler:             "file.handler",
			Runtime:             "runtime",
			Type:                "HTTP",
			Function:            "function",
			Checksum:            "sha256:78f9ac018e554365069108352dacabb7fbd15246edf19400677e3b54fe24e126",
			FunctionContentType: "text",
			Deps:                "dependencies",
			Topic:               "",
			Schedule:            "",
			Timeout:             "10",
			ServiceSpec: v1.ServiceSpec{
				Ports: []v1.ServicePort{
					{
						Name:       "function-port",
						Port:       int32(80),
						TargetPort: intstr.FromInt(80),
						NodePort:   0,
						Protocol:   v1.ProtocolTCP,
					},
				},
				Selector: map[string]string{
					"test":     "1",
					"function": "test",
				},
				Type:      v1.ServiceTypeClusterIP,
				ClusterIP: v1.ClusterIPNone,
			},
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Env: []v1.EnvVar{{
								Name:  "TEST",
								Value: "1",
							}},
							Resources: v1.ResourceRequirements{
								Limits: map[v1.ResourceName]resource.Quantity{
									v1.ResourceMemory: parsedMem,
								},
								Requests: map[v1.ResourceName]resource.Quantity{
									v1.ResourceMemory: parsedMem,
								},
							},
							Image: "test-image",
						},
					},
				},
			},
		},
	}
	if !reflect.DeepEqual(expectedFunction, *result) {
		t.Errorf("Unexpected result. Expecting:\n %+v\nReceived:\n %+v", expectedFunction, *result)
	}

	// It should take the default values
	result2, err := getFunctionDescription(fake.NewSimpleClientset(), "test", "default", "", "", "", "", "", "", "", "", "", false, nil, nil, []string{}, []string{}, expectedFunction)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedFunction, *result2) {
		t.Error("Unexpected result")
	}

	// Given parameters should take precedence from default values
	file, err = ioutil.TempFile("", "test")
	if err != nil {
		t.Error(err)
	}
	_, err = file.WriteString("function-modified")
	if err != nil {
		t.Error(err)
	}
	file.Close()
	defer os.Remove(file.Name()) // clean up
	input3Headless := false
	input3Port := int32(8080)
	result3, err := getFunctionDescription(fake.NewSimpleClientset(), "test", "default", "file.handler2", file.Name(), "dependencies2", "runtime2", "test_topic", "", "test-image2", "256Mi", "20", false, &input3Headless, &input3Port, []string{"TEST=2"}, []string{"test=2"}, expectedFunction)
	if err != nil {
		t.Error(err)
	}
	parsedMem2, _ := parseMemory("256Mi")
	newFunction := spec.Function{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Function",
			APIVersion: "k8s.io/v1",
		},
		Metadata: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
			Labels: map[string]string{
				"test": "2",
			},
		},
		Spec: spec.FunctionSpec{
			Handler:             "file.handler2",
			Runtime:             "runtime2",
			Type:                "PubSub",
			Function:            "function-modified",
			FunctionContentType: "text",
			Checksum:            "sha256:1958eb96d7d3cadedd0f327f09322eb7db296afb282ed91aa66cb4ab0dcc3c9f",
			Deps:                "dependencies2",
			Topic:               "test_topic",
			Schedule:            "",
			Timeout:             "20",
			ServiceSpec: v1.ServiceSpec{
				Ports: []v1.ServicePort{
					{
						Name:       "function-port",
						Port:       int32(8080),
						TargetPort: intstr.FromInt(8080),
						NodePort:   0,
						Protocol:   v1.ProtocolTCP,
					},
				},
				Selector: map[string]string{
					"test":     "2",
					"function": "test",
				},
				Type: v1.ServiceTypeClusterIP,
			},
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Env: []v1.EnvVar{{
								Name:  "TEST",
								Value: "2",
							}},
							Resources: v1.ResourceRequirements{
								Limits: map[v1.ResourceName]resource.Quantity{
									v1.ResourceMemory: parsedMem2,
								},
								Requests: map[v1.ResourceName]resource.Quantity{
									v1.ResourceMemory: parsedMem2,
								},
							},
							Image: "test-image2",
						},
					},
				},
			},
		},
	}
	if !reflect.DeepEqual(newFunction, *result3) {
		t.Errorf("Unexpected result. Expecting:\n %+v\n Received %+v\n", newFunction, *result3)
	}

	// It should detect that it is a Zip file
	file, err = os.Open(file.Name())
	if err != nil {
		t.Error(err)
	}
	newfile, err := os.Create(file.Name() + ".zip")
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(newfile.Name()) // clean up
	zipW := zip.NewWriter(newfile)
	info, err := file.Stat()
	if err != nil {
		t.Error(err)
	}
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		t.Error(err)
	}
	writer, err := zipW.CreateHeader(header)
	if err != nil {
		t.Error(err)
	}
	_, err = io.Copy(writer, file)
	if err != nil {
		t.Error(err)
	}
	file.Close()
	zipW.Close()
	result4, err := getFunctionDescription(fake.NewSimpleClientset(), "test", "default", "file.handler", newfile.Name(), "dependencies", "runtime", "", "", "", "", "", false, nil, nil, []string{}, []string{}, expectedFunction)
	if err != nil {
		t.Error(err)
	}
	if result4.Spec.FunctionContentType != "base64+zip" {
		t.Errorf("Should return base64+zip, received %s", result4.Spec.FunctionContentType)
	}
}
