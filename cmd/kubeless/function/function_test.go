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
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	kubelessApi "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/autoscaling/v2beta1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
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

func TestParseNodeSelectors(t *testing.T) {
	nodeSelectors := []string{
		"foo=bar",
		"baz:qux",
	}
	expected := map[string]string{
		"foo": "bar",
		"baz": "qux",
	}
	actual := parseNodeSelectors(nodeSelectors)
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

	result, err := getFunctionDescription("test", "default", "file.handler", file.Name(), "dependencies", "runtime", "test-image", "128Mi", "", "10", "Always", "serviceAccount", 8080, 0, false, []string{"TEST=1"}, []string{"test=1"}, []string{"secretName"}, []string{"foo1=bar1", "baz1:qux1"}, kubelessApi.Function{})

	if err != nil {
		t.Error(err)
	}
	parsedMem, _ := parseResource("128Mi")
	parsedCPU, _ := parseResource("")
	expectedFunction := kubelessApi.Function{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Function",
			APIVersion: "kubeless.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
			Labels: map[string]string{
				"test": "1",
			},
		},
		Spec: kubelessApi.FunctionSpec{
			Handler:             "file.handler",
			Runtime:             "runtime",
			Function:            "function",
			Checksum:            "sha256:78f9ac018e554365069108352dacabb7fbd15246edf19400677e3b54fe24e126",
			FunctionContentType: "text",
			Deps:                "dependencies",
			Timeout:             "10",
			Deployment: appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							ServiceAccountName: "serviceAccount",
							Containers: []v1.Container{
								{
									Env: []v1.EnvVar{{
										Name:  "TEST",
										Value: "1",
									}},
									Resources: v1.ResourceRequirements{
										Limits: map[v1.ResourceName]resource.Quantity{
											v1.ResourceMemory: parsedMem,
											v1.ResourceCPU:    parsedCPU,
										},
										Requests: map[v1.ResourceName]resource.Quantity{
											v1.ResourceMemory: parsedMem,
											v1.ResourceCPU:    parsedCPU,
										},
									},
									Image:           "test-image",
									ImagePullPolicy: v1.PullAlways,
									VolumeMounts: []v1.VolumeMount{
										{
											Name:      "secretName-vol",
											MountPath: "/secretName",
										},
									},
								},
							},
							Volumes: []v1.Volume{
								{
									Name: "secretName-vol",
									VolumeSource: v1.VolumeSource{
										Secret: &v1.SecretVolumeSource{
											SecretName: "secretName",
										},
									},
								},
							},
							NodeSelector: map[string]string{
								"foo1": "bar1",
								"baz1": "qux1",
							},
						},
					},
				},
			},
			ServiceSpec: v1.ServiceSpec{
				Ports: []v1.ServicePort{
					{Name: "http-function-port", Protocol: "TCP", Port: 8080, TargetPort: intstr.FromInt(8080)},
				},
				Selector: map[string]string{
					"test": "1",
				},
				Type: v1.ServiceTypeClusterIP,
			},
		},
	}
	if !reflect.DeepEqual(expectedFunction, *result) {
		t.Errorf("Unexpected result. Expecting:\n %+v\nReceived:\n %+v", expectedFunction, *result)
	}

	// It should take the default values
	result2, err := getFunctionDescription("test", "default", "", "", "", "", "", "", "", "", "Always", "", 8080, 0, false, []string{}, []string{}, []string{}, []string{}, expectedFunction)

	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedFunction, *result2) {
		t.Errorf("Unexpected result. Expecting:\n %+v\n Received %+v\n", expectedFunction, *result2)
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

	result3, err := getFunctionDescription("test", "default", "file.handler2", file.Name(), "dependencies2", "runtime2", "test-image2", "256Mi", "100m", "20", "Always", "NewServiceAccount", 8080, 0, false, []string{"TEST=2"}, []string{"test=2"}, []string{"secret2"}, []string{"foo2=bar2", "baz2:qux2"}, expectedFunction)

	if err != nil {
		t.Error(err)
	}
	parsedMem2, _ := parseResource("256Mi")
	parsedCPU2, _ := parseResource("100m")
	newFunction := kubelessApi.Function{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Function",
			APIVersion: "kubeless.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
			Labels: map[string]string{
				"test": "2",
			},
		},
		Spec: kubelessApi.FunctionSpec{
			Handler:             "file.handler2",
			Runtime:             "runtime2",
			Function:            "function-modified",
			FunctionContentType: "text",
			Checksum:            "sha256:1958eb96d7d3cadedd0f327f09322eb7db296afb282ed91aa66cb4ab0dcc3c9f",
			Deps:                "dependencies2",
			Timeout:             "20",
			Deployment: appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							ServiceAccountName: "NewServiceAccount",
							Containers: []v1.Container{
								{
									Env: []v1.EnvVar{{
										Name:  "TEST",
										Value: "2",
									}},
									Resources: v1.ResourceRequirements{
										Limits: map[v1.ResourceName]resource.Quantity{
											v1.ResourceMemory: parsedMem2,
											v1.ResourceCPU:    parsedCPU2,
										},
										Requests: map[v1.ResourceName]resource.Quantity{
											v1.ResourceMemory: parsedMem2,
											v1.ResourceCPU:    parsedCPU2,
										},
									},
									Image:           "test-image2",
									ImagePullPolicy: v1.PullAlways,
									VolumeMounts: []v1.VolumeMount{
										{
											Name:      "secretName-vol",
											MountPath: "/secretName",
										}, {
											Name:      "secret2-vol",
											MountPath: "/secret2",
										},
									},
								},
							},
							Volumes: []v1.Volume{
								{
									Name: "secretName-vol",
									VolumeSource: v1.VolumeSource{
										Secret: &v1.SecretVolumeSource{
											SecretName: "secretName",
										},
									},
								}, {
									Name: "secret2-vol",
									VolumeSource: v1.VolumeSource{
										Secret: &v1.SecretVolumeSource{
											SecretName: "secret2",
										},
									},
								},
							},
							NodeSelector: map[string]string{
								"foo2": "bar2",
								"baz2": "qux2",
							},
						},
					},
				},
			},
			ServiceSpec: v1.ServiceSpec{
				Ports: []v1.ServicePort{
					{Name: "http-function-port", Protocol: "TCP", Port: 8080, TargetPort: intstr.FromInt(8080)},
				},
				Selector: map[string]string{
					"test": "2",
				},
				Type: v1.ServiceTypeClusterIP,
			},
		},
	}
	if !reflect.DeepEqual(newFunction, *result3) {
		t.Errorf("Unexpected result. Expecting:\n %+v\n Received %+v\n", newFunction, *result3)
	}

	// It should detect that it is a Zip file or a compressed tar file
	file, err = os.Open(file.Name())
	if err != nil {
		t.Error(err)
	}

	zipFile, err := os.Create(file.Name() + ".zip")
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(zipFile.Name()) // clean up

	tarGzFile, err := os.Create(file.Name() + ".tar.gz")
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(tarGzFile.Name()) // clean up

	zipW := zip.NewWriter(zipFile)
	gzipW := gzip.NewWriter(tarGzFile)
	tarW := tar.NewWriter(gzipW)

	info, err := file.Stat()
	if err != nil {
		t.Error(err)
	}

	zipHeader, err := zip.FileInfoHeader(info)
	if err != nil {
		t.Error(err)
	}
	writer, err := zipW.CreateHeader(zipHeader)
	if err != nil {
		t.Error(err)
	}
	_, err = io.Copy(writer, file)
	if err != nil {
		t.Error(err)
	}

	tarHeader, err := tar.FileInfoHeader(info, info.Name())
	if err != nil {
		t.Error(err)
	}
	tarHeader.Name = file.Name()
	err = tarW.WriteHeader(tarHeader)
	if err != nil {
		t.Error(err)
	}
	_, err = io.Copy(writer, file)
	if err != nil {
		t.Error(err)
	}

	file.Close()
	zipW.Close()
	zipFile.Close()
	tarW.Close()
	gzipW.Close()
	tarGzFile.Close()

	result4A, err := getFunctionDescription("test", "default", "file.handler", zipFile.Name(), "dependencies", "runtime", "", "", "", "", "Always", "", 8080, 0, false, []string{}, []string{}, []string{}, []string{}, expectedFunction)
	if err != nil {
		t.Error(err)
	}
	if result4A.Spec.FunctionContentType != "base64+zip" {
		t.Errorf("Should return base64+zip, received %s", result4A.Spec.FunctionContentType)
	}

	result4B, err := getFunctionDescription("test", "default", "file.handler", tarGzFile.Name(), "dependencies", "runtime", "", "", "", "", "Always", "", 8080, 0, false, []string{}, []string{}, []string{}, []string{}, expectedFunction)
	if err != nil {
		t.Error(err)
	}
	if result4B.Spec.FunctionContentType != "base64+compressedtar" {
		t.Errorf("Should return base64+compressedtar, received %s", result4B.Spec.FunctionContentType)
	}

	// It should maintain previous HPA definition
	result5, err := getFunctionDescription("test", "default", "file.handler", file.Name(), "dependencies", "runtime", "test-image", "128Mi", "", "10", "Always", "serviceAccount", 8080, 0, false, []string{"TEST=1"}, []string{"test=1"}, []string{}, []string{}, kubelessApi.Function{

		Spec: kubelessApi.FunctionSpec{
			HorizontalPodAutoscaler: v2beta1.HorizontalPodAutoscaler{
				ObjectMeta: metav1.ObjectMeta{
					Name: "previous-hpa",
				},
			},
		},
	})
	if result5.Spec.HorizontalPodAutoscaler.ObjectMeta.Name != "previous-hpa" {
		t.Error("should maintain previous HPA definition")
	}

	// It should set the Port, ServicePort and headless service properly
	result6, err := getFunctionDescription("test", "default", "file.handler", file.Name(), "dependencies", "runtime", "test-image", "128Mi", "", "", "Always", "serviceAccount", 9091, 9092, true, []string{}, []string{}, []string{}, []string{}, kubelessApi.Function{})
	expectedPort := v1.ServicePort{
		Name:       "http-function-port",
		Port:       9092,
		TargetPort: intstr.FromInt(9091),
		NodePort:   0,
		Protocol:   v1.ProtocolTCP,
	}
	if !reflect.DeepEqual(result6.Spec.ServiceSpec.Ports[0], expectedPort) {
		t.Errorf("Unexpected port definition: %v", result6.Spec.ServiceSpec.Ports[0])
	}
	if result6.Spec.ServiceSpec.ClusterIP != v1.ClusterIPNone {
		t.Errorf("Unexpected clusterIP %v", result6.Spec.ServiceSpec.ClusterIP)
	}

	// it should create a function from a URL
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "function")
	}))
	defer ts.Close()

	expectedURLFunction := kubelessApi.Function{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Function",
			APIVersion: "kubeless.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
			Labels: map[string]string{
				"test": "1",
			},
		},
		Spec: kubelessApi.FunctionSpec{
			Handler:             "file.handler",
			Runtime:             "runtime",
			Function:            ts.URL,
			Checksum:            "sha256:78f9ac018e554365069108352dacabb7fbd15246edf19400677e3b54fe24e126",
			FunctionContentType: "url",
			Deps:                "dependencies",
			Timeout:             "10",
			Deployment: appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							ServiceAccountName: "serviceAccount",
							Containers: []v1.Container{
								{
									Env: []v1.EnvVar{{
										Name:  "TEST",
										Value: "1",
									}},
									Resources: v1.ResourceRequirements{
										Limits: map[v1.ResourceName]resource.Quantity{
											v1.ResourceMemory: parsedMem,
											v1.ResourceCPU:    parsedCPU,
										},
										Requests: map[v1.ResourceName]resource.Quantity{
											v1.ResourceMemory: parsedMem,
											v1.ResourceCPU:    parsedCPU,
										},
									},
									Image:           "test-image",
									ImagePullPolicy: v1.PullAlways,
									VolumeMounts: []v1.VolumeMount{
										{
											Name:      "secretName-vol",
											MountPath: "/secretName",
										},
									},
								},
							},
							Volumes: []v1.Volume{
								{
									Name: "secretName-vol",
									VolumeSource: v1.VolumeSource{
										Secret: &v1.SecretVolumeSource{
											SecretName: "secretName",
										},
									},
								},
							},
							NodeSelector: map[string]string{
								"foo3": "bar3",
								"baz3": "qux3",
							},
						},
					},
				},
			},
			ServiceSpec: v1.ServiceSpec{
				Ports: []v1.ServicePort{
					{Name: "http-function-port", Protocol: "TCP", Port: 8080, TargetPort: intstr.FromInt(8080)},
				},
				Selector: map[string]string{
					"test": "1",
				},
				Type: v1.ServiceTypeClusterIP,
			},
		},
	}

	result7, err := getFunctionDescription("test", "default", "file.handler", ts.URL, "dependencies", "runtime", "test-image", "128Mi", "", "10", "Always", "serviceAccount", 8080, 0, false, []string{"TEST=1"}, []string{"test=1"}, []string{"secretName"}, []string{"foo3=bar3", "baz3:qux3"}, kubelessApi.Function{})

	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(expectedURLFunction, *result7) {
		t.Errorf("Unexpected result. Expecting:\n %+v\nReceived:\n %+v", expectedURLFunction, *result7)
	}

	// It should handle zip files and compressed tar files from a URL and detect url+zip and url+compressedtar encoding respectively
	zipBytes, err := ioutil.ReadFile(zipFile.Name())
	if err != nil {
		t.Error(err)
	}

	ts2A := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(zipBytes)
	}))
	defer ts2A.Close()

	expectedURLFunction.Spec.FunctionContentType = "url+zip"
	expectedURLFunction.Spec.Function = ts2A.URL + "/test.zip"
	expectedURLFunction.Spec.Checksum, err = getSha256(zipBytes)
	if err != nil {
		t.Error(err)
	}

	result8A, err := getFunctionDescription("test", "default", "file.handler", ts2A.URL+"/test.zip", "dependencies", "runtime", "test-image", "128Mi", "", "10", "Always", "serviceAccount", 8080, 0, false, []string{"TEST=1"}, []string{"test=1"}, []string{"secretName"}, []string{"foo3=bar3", "baz3:qux3"}, kubelessApi.Function{})
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedURLFunction, *result8A) {
		t.Errorf("Unexpected result. Expecting:\n %+v\nReceived:\n %+v", expectedURLFunction, *result8A)
	}

	tarGzBytes, err := ioutil.ReadFile(tarGzFile.Name())
	if err != nil {
		t.Error(err)
	}
	ts2B := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(tarGzBytes)
	}))
	defer ts2B.Close()

	expectedURLFunction.Spec.FunctionContentType = "url+compressedtar"
	expectedURLFunction.Spec.Function = ts2B.URL + "/test.tar.gz"
	expectedURLFunction.Spec.Checksum, err = getSha256(tarGzBytes)
	if err != nil {
		t.Error(err)
	}

	result8B, err := getFunctionDescription("test", "default", "file.handler", ts2B.URL+"/test.tar.gz", "dependencies", "runtime", "test-image", "128Mi", "", "10", "Always", "serviceAccount", 8080, 0, false, []string{"TEST=1"}, []string{"test=1"}, []string{"secretName"}, []string{"foo3=bar3", "baz3:qux3"}, kubelessApi.Function{})
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedURLFunction, *result8B) {
		t.Errorf("Unexpected result. Expecting:\n %+v\nReceived:\n %+v", expectedURLFunction, *result8B)
	}
	// end test
}

func getSha256(bytes []byte) (string, error) {
	h := sha256.New()
	_, err := h.Write(bytes)
	if err != nil {
		return "", err
	}
	checksum := hex.EncodeToString(h.Sum(nil))
	return "sha256:" + checksum, nil
}
