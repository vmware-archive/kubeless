package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/kubeless/kubeless/pkg/spec"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apimachinery"
	"k8s.io/apimachinery/pkg/apimachinery/registered"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/autoscaling/v2alpha1"
	av2alpha1 "k8s.io/client-go/pkg/apis/autoscaling/v2alpha1"
	batchv2alpha1 "k8s.io/client-go/pkg/apis/batch/v2alpha1"
	xv1beta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/rest"
	restFake "k8s.io/client-go/rest/fake"
	ktesting "k8s.io/client-go/testing"
)

func getEnvValueFromList(envName string, l []v1.EnvVar) string {
	var res v1.EnvVar
	for _, env := range l {
		if env.Name == envName {
			res = env
			break
		}
	}
	return res.Value
}

func TestEnsureConfigMap(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	or := []metav1.OwnerReference{
		{
			Kind:       "Function",
			APIVersion: "k8s.io",
		},
	}
	ns := "default"
	funcLabels := map[string]string{
		"foo": "bar",
	}
	f1Name := "f1"
	f1 := &spec.Function{
		Metadata: metav1.ObjectMeta{
			Name:      f1Name,
			Namespace: ns,
			Labels:    funcLabels,
		},
		Spec: spec.FunctionSpec{
			Function: "function",
			Deps:     "deps",
			Handler:  "foo.bar",
			Runtime:  "python2.7",
		},
	}
	err := EnsureFuncConfigMap(clientset, f1, or)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	cm, err := clientset.CoreV1().ConfigMaps(ns).Get(f1Name, metav1.GetOptions{})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	expectedCM := v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            f1Name,
			Namespace:       ns,
			Labels:          funcLabels,
			OwnerReferences: or,
		},
		Data: map[string]string{
			"handler":          "foo.bar",
			"foo.py":           "function",
			"requirements.txt": "deps",
		},
	}
	if !reflect.DeepEqual(*cm, expectedCM) {
		t.Errorf("Unexpected ConfigMap:\n %+v\nExpecting:\n %+v", *cm, expectedCM)
	}

	// It should skip the dependencies field in case it is not supported
	f2Name := "f2"
	f2 := &spec.Function{
		Metadata: metav1.ObjectMeta{
			Name:      f2Name,
			Namespace: ns,
		},
		Spec: spec.FunctionSpec{
			Function: "function",
			Handler:  "foo.bar",
			Runtime:  "cobol",
		},
	}
	err = EnsureFuncConfigMap(clientset, f2, or)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	cm, err = clientset.CoreV1().ConfigMaps(ns).Get(f2Name, metav1.GetOptions{})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	expectedData := map[string]string{
		"handler": "foo.bar",
		"foo":     "function",
	}
	if !reflect.DeepEqual(cm.Data, expectedData) {
		t.Errorf("Unexpected ConfigMap:\n %+v\nExpecting:\n %+v", cm.Data, expectedData)
	}

	// If there is already a config map it should update the previous one
	f2 = &spec.Function{
		Metadata: metav1.ObjectMeta{
			Name:      f2Name,
			Namespace: ns,
		},
		Spec: spec.FunctionSpec{
			Function: "function2",
			Handler:  "foo2.bar2",
			Runtime:  "python3.4",
		},
	}
	err = EnsureFuncConfigMap(clientset, f2, or)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	cm, err = clientset.CoreV1().ConfigMaps(ns).Get(f2Name, metav1.GetOptions{})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	expectedData = map[string]string{
		"handler":          "foo2.bar2",
		"foo2.py":          "function2",
		"requirements.txt": "",
	}
	if !reflect.DeepEqual(cm.Data, expectedData) {
		t.Errorf("Unexpected ConfigMap:\n %+v\nExpecting:\n %+v", cm.Data, expectedData)
	}
}

func TestEnsureService(t *testing.T) {
	fakeSvc := v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "myns",
			Name:      "foo",
		},
	}
	clientset := fake.NewSimpleClientset(&fakeSvc)
	or := []metav1.OwnerReference{
		{
			Kind:       "Function",
			APIVersion: "k8s.io",
		},
	}
	ns := "default"
	funcLabels := map[string]string{
		"foo": "bar",
	}
	f1Name := "f1"
	f1 := &spec.Function{
		Metadata: metav1.ObjectMeta{
			Name:      f1Name,
			Namespace: ns,
			Labels:    funcLabels,
		},
		Spec: spec.FunctionSpec{
			Function: "function",
			Deps:     "deps",
			Handler:  "foo.bar",
			Runtime:  "python2.7",
			ServiceSpec: v1.ServiceSpec{
				Ports: []v1.ServicePort{
					{
						Name:       "function-port",
						Port:       8080,
						TargetPort: intstr.FromInt(8080),
						NodePort:   0,
						Protocol:   v1.ProtocolTCP,
					},
				},
				Selector: funcLabels,
				Type:     v1.ServiceTypeClusterIP,
			},
		},
	}
	err := EnsureFuncService(clientset, f1, or)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	svc, err := clientset.CoreV1().Services(ns).Get(f1Name, metav1.GetOptions{})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	expectedSVC := v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            f1Name,
			Namespace:       ns,
			Labels:          funcLabels,
			OwnerReferences: or,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:       "function-port",
					Port:       8080,
					TargetPort: intstr.FromInt(8080),
					NodePort:   0,
					Protocol:   v1.ProtocolTCP,
				},
			},
			Selector: funcLabels,
			Type:     v1.ServiceTypeClusterIP,
		},
	}
	if !reflect.DeepEqual(*svc, expectedSVC) {
		t.Errorf("Unexpected service:\n %+v\nExpecting:\n %+v", *svc, expectedSVC)
	}

	// If there is already a service it should update the previous one
	newLabels := map[string]string{
		"foobar": "barfoo",
	}
	f1.Metadata.Labels = newLabels
	err = EnsureFuncService(clientset, f1, or)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	svc, err = clientset.CoreV1().Services(ns).Get(f1Name, metav1.GetOptions{})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if !reflect.DeepEqual(svc.ObjectMeta.Labels, newLabels) {
		t.Error("Unable to update the service")
	}
}

func TestEnsureDeployment(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	or := []metav1.OwnerReference{
		{
			Kind:       "Function",
			APIVersion: "k8s.io",
		},
	}
	ns := "default"
	funcLabels := map[string]string{
		"foo": "bar",
	}
	funcAnno := map[string]string{
		"bar": "foo",
	}
	f1Name := "f1"
	f1Port := int32(8080)
	f1 := &spec.Function{
		Metadata: metav1.ObjectMeta{
			Name:      f1Name,
			Namespace: ns,
			Labels:    funcLabels,
		},
		Spec: spec.FunctionSpec{
			Function: "function",
			Deps:     "deps",
			Handler:  "foo.bar",
			Runtime:  "python2.7",
			ServiceSpec: v1.ServiceSpec{
				Ports: []v1.ServicePort{
					{
						Name:       "function-port",
						Port:       f1Port,
						TargetPort: intstr.FromInt(int(f1Port)),
						NodePort:   0,
						Protocol:   v1.ProtocolTCP,
					},
				},
				Selector: funcLabels,
				Type:     v1.ServiceTypeClusterIP,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: funcAnno,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Env: []v1.EnvVar{
								{
									Name:  "foo",
									Value: "bar",
								},
							},
						},
					},
				},
			},
		},
	}
	// Testing happy path
	err := EnsureFuncDeployment(clientset, f1, or)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	dpm, err := clientset.ExtensionsV1beta1().Deployments(ns).Get(f1Name, metav1.GetOptions{})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	expectedObjectMeta := metav1.ObjectMeta{
		Name:            f1Name,
		Namespace:       ns,
		Labels:          funcLabels,
		OwnerReferences: or,
	}
	if !reflect.DeepEqual(dpm.ObjectMeta, expectedObjectMeta) {
		t.Errorf("Unable to set metadata. Received:\n %+v\nExpecting:\n %+v", dpm.ObjectMeta, expectedObjectMeta)
	}
	expectedAnnotations := map[string]string{
		"prometheus.io/scrape": "true",
		"prometheus.io/path":   "/metrics",
		"prometheus.io/port":   strconv.Itoa(int(f1Port)),
		"bar":                  "foo",
	}
	for i := range expectedAnnotations {
		if dpm.Spec.Template.Annotations[i] != expectedAnnotations[i] {
			t.Errorf("Expecting annotation %s but received %s", expectedAnnotations[i], dpm.Spec.Template.Annotations[i])
		}
	}
	if dpm.Spec.Template.Annotations["bar"] != "foo" {
		t.Error("Unable to set annotations")
	}
	expectedContainer := v1.Container{
		Name:  f1Name,
		Image: "kubeless/python@sha256:0f3b64b654df5326198e481cd26e73ecccd905aae60810fc9baea4dcbb61f697",
		Ports: []v1.ContainerPort{
			{
				ContainerPort: int32(f1Port),
			},
		},
		Env: []v1.EnvVar{
			{
				Name:  "foo",
				Value: "bar",
			},
			{
				Name:  "FUNC_HANDLER",
				Value: "bar",
			},
			{
				Name:  "MOD_NAME",
				Value: "foo",
			},
			{
				Name:  "FUNC_TIMEOUT",
				Value: "180",
			},
			{
				Name:  "FUNC_PORT",
				Value: strconv.Itoa(int(f1Port)),
			},
			{
				Name:  "TOPIC_NAME",
				Value: "",
			},
			{
				Name:  "PYTHONPATH",
				Value: "/kubeless/lib/python2.7/site-packages",
			},
		},
		VolumeMounts: []v1.VolumeMount{
			{
				Name:      f1Name,
				MountPath: "/kubeless",
			},
		},
		LivenessProbe: &v1.Probe{
			InitialDelaySeconds: int32(3),
			PeriodSeconds:       int32(30),
			Handler: v1.Handler{
				HTTPGet: &v1.HTTPGetAction{
					Path: "/healthz",
					Port: intstr.FromInt(int(f1Port)),
				},
			},
		},
	}
	if !reflect.DeepEqual(dpm.Spec.Template.Spec.Containers[0], expectedContainer) {
		t.Errorf("Unexpected container definition. Received:\n %+v\nExpecting:\n %+v", dpm.Spec.Template.Spec.Containers[0], expectedContainer)
	}
	// Init containers behavior should be tested with integration tests
	if len(dpm.Spec.Template.Spec.InitContainers) < 1 {
		t.Errorf("Expecting at least an init container to install deps")
	}

	// If no handler and function is given it should not fail
	f2 := spec.Function{}
	f2 = *f1
	f2.Metadata.Name = "func2"
	f2.Spec.Function = ""
	f2.Spec.Handler = ""
	err = EnsureFuncDeployment(clientset, &f2, or)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	dpm, err = clientset.ExtensionsV1beta1().Deployments(ns).Get("func2", metav1.GetOptions{})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	// If the Image has been already provided it should not resolve it
	f3 := spec.Function{}
	f3 = *f1
	f3.Metadata.Name = "func3"
	f3.Spec.Template.Spec.Containers[0].Image = "test-image"
	err = EnsureFuncDeployment(clientset, &f3, or)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	dpm, err = clientset.ExtensionsV1beta1().Deployments(ns).Get("func3", metav1.GetOptions{})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if dpm.Spec.Template.Spec.Containers[0].Image != "test-image" {
		t.Errorf("Unexpected Image Name: %s", dpm.Spec.Template.Spec.Containers[0].Image)
	}

	// If no function is given it should not use an init container
	f4 := spec.Function{}
	f4 = *f1
	f4.Metadata.Name = "func4"
	f4.Spec.Function = ""
	f4.Spec.Deps = ""
	err = EnsureFuncDeployment(clientset, &f4, or)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	dpm, err = clientset.ExtensionsV1beta1().Deployments(ns).Get("func4", metav1.GetOptions{})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if len(dpm.Spec.Template.Spec.InitContainers) > 0 {
		t.Error("It should not setup an init container")
	}

	// If the function is the type PubSub it should not contain a livenessProbe
	f5 := spec.Function{}
	f5 = *f1
	f5.Metadata.Name = "func5"
	f5.Spec.Type = "PubSub"
	err = EnsureFuncDeployment(clientset, &f5, or)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	dpm, err = clientset.ExtensionsV1beta1().Deployments(ns).Get("func5", metav1.GetOptions{})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if dpm.Spec.Template.Spec.Containers[0].LivenessProbe != nil {
		t.Error("It should not setup a liveness probe")
	}

	// It should update a deployment if it is already present
	f6 := spec.Function{}
	f6 = *f1
	f6.Spec.Handler = "foo.bar2"
	err = EnsureFuncDeployment(clientset, &f6, or)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	dpm, err = clientset.ExtensionsV1beta1().Deployments(ns).Get(f1Name, metav1.GetOptions{})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if getEnvValueFromList("FUNC_HANDLER", dpm.Spec.Template.Spec.Containers[0].Env) != "bar2" {
		t.Error("Unable to update deployment")
	}

	// It should return an error if some dependencies are given but the runtime is not supported
	f7 := spec.Function{}
	f7 = *f1
	f7.Metadata.Name = "func7"
	f7.Spec.Deps = "deps"
	f7.Spec.Runtime = "cobol"
	err = EnsureFuncDeployment(clientset, &f7, or)
	if err == nil {
		t.Errorf("An error should be thrown")
	}

	// If a timeout is specified it should set an environment variable FUNC_TIMEOUT
	f8 := spec.Function{}
	f8 = *f1
	f8.Metadata.Name = "func8"
	f8.Spec.Timeout = "10"
	err = EnsureFuncDeployment(clientset, &f8, or)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	dpm, err = clientset.ExtensionsV1beta1().Deployments(ns).Get("func8", metav1.GetOptions{})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if getEnvValueFromList("FUNC_TIMEOUT", dpm.Spec.Template.Spec.Containers[0].Env) != "10" {
		t.Error("Unable to set timeout")
	}
}

func fakeRESTClient(f func(req *http.Request) (*http.Response, error)) *restFake.RESTClient {
	reg := registered.NewOrDie("v1")
	legacySchema := schema.GroupVersion{
		Group:   "",
		Version: "v1",
	}
	newSchema := schema.GroupVersion{
		Group:   "k8s.io",
		Version: "v1",
	}
	reg.RegisterGroup(apimachinery.GroupMeta{
		GroupVersion: legacySchema,
	})
	reg.RegisterGroup(apimachinery.GroupMeta{
		GroupVersion: newSchema,
	})
	return &restFake.RESTClient{
		APIRegistry:          reg,
		NegotiatedSerializer: api.Codecs,
		Client:               restFake.CreateHTTPClient(f),
	}
}

func objBody(object interface{}) io.ReadCloser {
	output, err := json.Marshal(object)
	if err != nil {
		panic(err)
	}
	return ioutil.NopCloser(bytes.NewReader([]byte(output)))
}

func TestEnsureCronJob(t *testing.T) {
	or := []metav1.OwnerReference{
		{
			Kind:       "Function",
			APIVersion: "k8s.io",
		},
	}
	ns := "default"
	f1Name := "func1"
	f1 := &spec.Function{
		Metadata: metav1.ObjectMeta{
			Name:      f1Name,
			Namespace: ns,
		},
		Spec: spec.FunctionSpec{
			Timeout:  "120",
			Schedule: "*/10 * * * *",
		},
	}

	expectedMeta := metav1.ObjectMeta{
		Name:            "trigger-" + f1Name,
		Namespace:       ns,
		OwnerReferences: or,
	}

	client := fakeRESTClient(func(req *http.Request) (*http.Response, error) {
		header := http.Header{}
		header.Set("Content-Type", runtime.ContentTypeJSON)
		listObj := batchv2alpha1.CronJobList{}
		if req.Method == "POST" {
			reqCronJobBytes, err := ioutil.ReadAll(req.Body)
			if err != nil {
				t.Fatal(err)
			}
			cronJob := batchv2alpha1.CronJob{}
			err = json.Unmarshal(reqCronJobBytes, &cronJob)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(expectedMeta, cronJob.ObjectMeta) {
				t.Errorf("Unexpected metadata metadata. Expecting\n%+v \nReceived:\n%+v", expectedMeta, cronJob.ObjectMeta)
			}
			if *cronJob.Spec.SuccessfulJobsHistoryLimit != int32(3) {
				t.Errorf("Unexpected SuccessfulJobsHistoryLimit: %d", *cronJob.Spec.SuccessfulJobsHistoryLimit)
			}
			if *cronJob.Spec.FailedJobsHistoryLimit != int32(1) {
				t.Errorf("Unexpected FailedJobsHistoryLimit: %d", *cronJob.Spec.FailedJobsHistoryLimit)
			}
			if *cronJob.Spec.JobTemplate.Spec.ActiveDeadlineSeconds != int64(120) {
				t.Errorf("Unexpected ActiveDeadlineSeconds: %d", *cronJob.Spec.JobTemplate.Spec.ActiveDeadlineSeconds)
			}
			expectedCommand := []string{"wget", "-qO-", fmt.Sprintf("http://%s.%s.svc.cluster.local:8080", f1Name, ns)}
			if !reflect.DeepEqual(cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Args, expectedCommand) {
				t.Errorf("Unexpected command %s", strings.Join(cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Args, " "))
			}
		} else {
			t.Fatalf("unexpected verb %s", req.Method)
		}
		switch req.URL.Path {
		case "/apis/batch/v2alpha1/namespaces/default/cronjobs":
			return &http.Response{
				StatusCode: 200,
				Header:     header,
				Body:       objBody(&listObj),
			}, nil
		default:
			t.Fatalf("unexpected request: %#v\n%#v", req.URL, req)
			return nil, nil
		}
	})
	err := EnsureFuncCronJob(client, f1, or, "batch/v2alpha1")
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	// It should update the existing cronJob if it is already created
	updateCalled := false
	client = fakeRESTClient(func(req *http.Request) (*http.Response, error) {
		header := http.Header{}
		header.Set("Content-Type", runtime.ContentTypeJSON)
		switch req.Method {
		case "POST":
			return &http.Response{
				StatusCode: http.StatusConflict,
				Header:     header,
				Body:       objBody(nil),
			}, nil
		case "GET":
			previousCronJob := batchv2alpha1.CronJob{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "123456",
				},
			}
			return &http.Response{
				StatusCode: 200,
				Header:     header,
				Body:       objBody(&previousCronJob),
			}, nil
		case "PUT":
			updateCalled = true
			reqCronJobBytes, err := ioutil.ReadAll(req.Body)
			if err != nil {
				t.Fatal(err)
			}
			cronJob := batchv2alpha1.CronJob{}
			err = json.Unmarshal(reqCronJobBytes, &cronJob)
			if err != nil {
				t.Fatal(err)
			}
			if cronJob.ObjectMeta.ResourceVersion != "123456" {
				t.Error("Expecting that the object to update contains the previous information")
			}
			listObj := batchv2alpha1.CronJobList{}
			return &http.Response{
				StatusCode: 200,
				Header:     header,
				Body:       objBody(&listObj),
			}, nil
		default:
			t.Fatalf("unexpected request: %#v\n%#v", req.URL, req)
			return nil, nil
		}
	})
	err = EnsureFuncCronJob(client, f1, or, "batch/v2alpha1")
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if !updateCalled {
		t.Errorf("Expect the update method to be called")
	}

	// IT should change the endpoint
	client = fakeRESTClient(func(req *http.Request) (*http.Response, error) {
		header := http.Header{}
		header.Set("Content-Type", runtime.ContentTypeJSON)
		if req.URL.Path != "/apis/batch/v1beta1/namespaces/default/cronjobs" {
			t.Errorf("Unexpected URL %s", req.URL.Path)
		}
		return &http.Response{
			StatusCode: 200,
			Header:     header,
			Body:       objBody(nil),
		}, nil
	})
	err = EnsureFuncCronJob(client, f1, or, "batch/v1beta1")
}

func doesNotContain(envs []v1.EnvVar, env v1.EnvVar) bool {
	for _, e := range envs {
		if e == env {
			return false
		}
	}
	return true
}

func TestCreateIngressResource(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	f1 := &spec.Function{
		Metadata: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "myns",
			UID:       "1234",
		},
		Spec: spec.FunctionSpec{
			ServiceSpec: v1.ServiceSpec{
				Ports: []v1.ServicePort{
					{
						TargetPort: intstr.FromInt(8080),
					},
				},
			},
		},
	}
	if err := CreateIngress(clientset, f1, "bar", "foo.bar", "myns", false); err != nil {
		t.Fatalf("Creating ingress returned err: %v", err)
	}
	if err := CreateIngress(clientset, f1, "bar", "foo.bar", "myns", false); err != nil {
		if !k8sErrors.IsAlreadyExists(err) {
			t.Fatalf("Expect object is already exists, got %v", err)
		}
	}
	f1.Spec.ServiceSpec.Ports = []v1.ServicePort{}
	if err := CreateIngress(clientset, f1, "bar", "foo.bar", "myns", false); err == nil {
		t.Fatal("Expect create ingress fails, got success")
	}
}

func TestCreateIngressResourceWithTLSAcme(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	f1 := &spec.Function{
		Metadata: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "myns",
			UID:       "1234",
		},
		Spec: spec.FunctionSpec{
			ServiceSpec: v1.ServiceSpec{
				Ports: []v1.ServicePort{
					{
						TargetPort: intstr.FromInt(8080),
					},
				},
			},
		},
	}

	if err := CreateIngress(clientset, f1, "foo", "foo.bar", "myns", true); err != nil {
		t.Fatalf("Creating ingress returned err: %v", err)
	}

	ingress, err := clientset.ExtensionsV1beta1().Ingresses("myns").Get("foo", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Getting Ingress returned err: %v", err)
	}

	annotations := ingress.ObjectMeta.Annotations
	if annotations == nil || len(annotations) == 0 ||
		annotations["kubernetes.io/tls-acme"] != "true" ||
		annotations["ingress.kubernetes.io/ssl-redirect"] != "true" {
		t.Fatal("Missing or wrong annotations!")
	}

	tls := ingress.Spec.TLS
	if tls == nil || len(tls) != 1 ||
		tls[0].SecretName == "" ||
		tls[0].Hosts == nil || len(tls[0].Hosts) != 1 || tls[0].Hosts[0] == "" {
		t.Fatal("Missing or incomplete TLS spec!")
	}
}

func TestDeleteIngressResource(t *testing.T) {
	myNsFoo := metav1.ObjectMeta{
		Namespace: "myns",
		Name:      "foo",
	}

	ing := xv1beta1.Ingress{
		ObjectMeta: myNsFoo,
	}

	clientset := fake.NewSimpleClientset(&ing)
	if err := DeleteIngress(clientset, "foo", "myns"); err != nil {
		t.Fatalf("Deleting ingress returned err: %v", err)
	}
	a := clientset.Actions()
	if ns := a[0].GetNamespace(); ns != "myns" {
		t.Errorf("deleted ingress from wrong namespace (%s)", ns)
	}
	if name := a[0].(ktesting.DeleteAction).GetName(); name != "foo" {
		t.Errorf("deleted ingress with wrong name (%s)", name)
	}
}

func fakeConfig() *rest.Config {
	return &rest.Config{
		Host: "https://example.com:443",
		ContentConfig: rest.ContentConfig{
			GroupVersion: &schema.GroupVersion{
				Group:   "",
				Version: "v1",
			},
			NegotiatedSerializer: api.Codecs,
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
	hpaDef := v2alpha1.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
	}
	if err := CreateAutoscale(clientset, hpaDef); err != nil {
		t.Fatalf("Creating autoscale returned err: %v", err)
	}

	hpa, err := clientset.AutoscalingV2alpha1().HorizontalPodAutoscalers(ns).Get(name, metav1.GetOptions{})
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

	as := av2alpha1.HorizontalPodAutoscaler{
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

func TestGetProvisionContainer(t *testing.T) {
	rvol := v1.VolumeMount{Name: "runtime", MountPath: "/runtime"}
	dvol := v1.VolumeMount{Name: "deps", MountPath: "/deps"}
	c, err := getProvisionContainer("test", "sha256:abc1234", "test.func", "test.foo", "text", "python2.7", rvol, dvol)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	expectedContainer := v1.Container{
		Name:            "prepare",
		Image:           "kubeless/unzip@sha256:f162c062973cca05459834de6ed14c039d45df8cdb76097f50b028a1621b3697",
		Command:         []string{"sh", "-c"},
		Args:            []string{"echo 'abc1234  /deps/test.func' > /deps/test.func.sha256 && sha256sum -c /deps/test.func.sha256 && cp /deps/test.func /runtime/test.py && cp /deps/requirements.txt /runtime"},
		VolumeMounts:    []v1.VolumeMount{rvol, dvol},
		ImagePullPolicy: v1.PullIfNotPresent,
	}
	if !reflect.DeepEqual(expectedContainer, c) {
		t.Errorf("Unexpected result:\n %+v", c)
	}

	// If the content type is encoded it should decode it
	c, err = getProvisionContainer("Zm9vYmFyCg==", "sha256:abc1234", "test.func", "test.foo", "base64", "python2.7", rvol, dvol)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if !strings.HasPrefix(c.Args[0], "base64 -d < /deps/test.func > /deps/test.func.decoded") {
		t.Errorf("Unexpected command: %s", c.Args[0])
	}

	// It should skip the dependencies installation if the runtime is not supported
	c, err = getProvisionContainer("function", "sha256:abc1234", "test.func", "test.foo", "text", "cobol", rvol, dvol)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if strings.Contains(c.Args[0], "cp /deps ") {
		t.Errorf("Unexpected command: %s", c.Args[0])
	}

	// It should extract the file in case it is a Zip
	c, err = getProvisionContainer("Zm9vYmFyCg==", "sha256:abc1234", "test.zip", "test.foo", "base64+zip", "python2.7", rvol, dvol)
	if !strings.Contains(c.Args[0], "unzip -o /deps/test.zip.decoded -d /runtime") {
		t.Errorf("Unexpected command: %s", c.Args[0])
	}

}

func TestServiceSpec(t *testing.T) {
	f1 := &spec.Function{
		Metadata: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "myns",
			Labels: map[string]string{
				"function": "foo",
			},
		},
		Spec: spec.FunctionSpec{
			ServiceSpec: v1.ServiceSpec{
				Ports: []v1.ServicePort{
					{
						TargetPort: intstr.FromInt(9000),
					},
				},
				Selector: map[string]string{
					"function": "foo",
				},
			},
		},
	}

	eSvc := v1.ServiceSpec{
		Ports: []v1.ServicePort{
			{
				Name:       "function-port",
				Protocol:   v1.ProtocolTCP,
				Port:       8080,
				TargetPort: intstr.FromInt(8080),
			},
		},
		Selector: map[string]string{
			"function": "foo",
		},
		Type: v1.ServiceTypeClusterIP,
	}

	aSvc := serviceSpec(f1)
	if !reflect.DeepEqual(f1.Spec.ServiceSpec, aSvc) {
		t.Errorf("Unexpected result:\n %+v", aSvc)
	}

	f1.Spec.ServiceSpec = v1.ServiceSpec{}
	aSvc = serviceSpec(f1)
	if !reflect.DeepEqual(aSvc, eSvc) {
		t.Errorf("Unexpected result:\n %+v", aSvc)
	}
}
