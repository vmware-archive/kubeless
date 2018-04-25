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

	kubelessApi "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	"github.com/kubeless/kubeless/pkg/langruntime"

	v2beta1 "k8s.io/api/autoscaling/v2beta1"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	xv1beta1 "k8s.io/api/extensions/v1beta1"
	extensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	fakeextensionsapi "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apimachinery"
	"k8s.io/apimachinery/pkg/apimachinery/registered"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
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
			APIVersion: "kubeless.io/v1beta1",
		},
	}
	ns := "default"
	funcLabels := map[string]string{
		"foo": "bar",
	}
	f1Name := "f1"
	f1 := &kubelessApi.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name:      f1Name,
			Namespace: ns,
			Labels:    funcLabels,
		},
		Spec: kubelessApi.FunctionSpec{
			Function: "function",
			Deps:     "deps",
			Handler:  "foo.bar",
			Runtime:  "python2.7",
		},
	}

	langruntime.AddFakeConfig(clientset)
	lr := langruntime.SetupLangRuntime(clientset)
	lr.ReadConfigMap()

	err := EnsureFuncConfigMap(clientset, f1, or, lr)
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
	f2 := &kubelessApi.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name:      f2Name,
			Namespace: ns,
		},
		Spec: kubelessApi.FunctionSpec{
			Function: "function",
			Handler:  "foo.bar",
			Runtime:  "cobol",
		},
	}

	err = EnsureFuncConfigMap(clientset, f2, or, lr)
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
	f2 = &kubelessApi.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name:      f2Name,
			Namespace: ns,
		},
		Spec: kubelessApi.FunctionSpec{
			Function: "function2",
			Handler:  "foo2.bar2",
			Runtime:  "python3.4",
		},
	}
	err = EnsureFuncConfigMap(clientset, f2, or, lr)
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
			APIVersion: "kubeless.io/v1beta1",
		},
	}
	ns := "default"
	funcLabels := map[string]string{
		"foo": "bar",
	}
	f1Name := "f1"
	f1 := &kubelessApi.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name:      f1Name,
			Namespace: ns,
			Labels:    funcLabels,
		},
		Spec: kubelessApi.FunctionSpec{
			Function: "function",
			Deps:     "deps",
			Handler:  "foo.bar",
			Runtime:  "python2.7",
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
					Name:       "http-function-port",
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
	f1.ObjectMeta.Labels = newLabels
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
	if reflect.DeepEqual(svc.Spec.Selector, newLabels) {
		t.Error("It should not update the selector")
	}
}

func TestEnsureImage(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	langruntime.AddFakeConfig(clientset)
	lr := langruntime.SetupLangRuntime(clientset)
	lr.ReadConfigMap()
	ns := "default"
	f1Name := "f1"
	or := []metav1.OwnerReference{
		{
			Kind:       "Function",
			APIVersion: "kubeless.io/v1beta1",
		},
	}
	f1 := &kubelessApi.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name:      f1Name,
			Namespace: ns,
		},
		Spec: kubelessApi.FunctionSpec{
			Function: "function",
			Deps:     "deps",
			Handler:  "foo.bar",
			Runtime:  "python2.7",
		},
	}
	// Testing happy path
	pullSecrets := []v1.LocalObjectReference{
		{Name: "creds"},
	}
	err := EnsureFuncImage(clientset, f1, lr, or, "user/image", "4840d87600137157493ba43a24f0b4bb6cf524ebbf095ce96c79f85bf5a3ff5a", "kubeless/builder", "registry.docker.io", "registry-creds", "unzip", true, pullSecrets)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	jobs, err := clientset.BatchV1().Jobs(ns).List(metav1.ListOptions{})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if len(jobs.Items) != 1 {
		t.Errorf("It should have created the build job")
	}
	buildContainer := jobs.Items[0].Spec.Template.Spec.Containers[0]
	if buildContainer.Image != "kubeless/builder" {
		t.Errorf("Image %s of build job is not recognised", jobs.Items[0].Spec.Template.Spec.Containers[0].Image)
	}
	dockerConfigFolder := ""
	for _, envvar := range buildContainer.Env {
		if envvar.Name == "DOCKER_CONFIG_FOLDER" {
			dockerConfigFolder = envvar.Value
		}
	}
	if dockerConfigFolder == "" {
		t.Error("Builder image relies on the env var DOCKER_CONFIG_FOLDER to authenticate")
	}
	initContainer := jobs.Items[0].Spec.Template.Spec.InitContainers[0]
	if initContainer.Image != "unzip" {
		t.Errorf("Unexpected init image %s", initContainer.Image)
	}
	if reflect.DeepEqual(jobs.Items[0].Spec.Template.Spec.ImagePullSecrets, pullSecrets) {
		t.Error("Missing ImagePullSecrets")
	}
}

func getDefaultFunc(name, ns string) *kubelessApi.Function {
	fPort := int32(8080)
	f := kubelessApi.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: kubelessApi.FunctionSpec{
			Function: "function",
			Deps:     "deps",
			Handler:  "foo.bar",
			Runtime:  "python2.7",
			ServiceSpec: v1.ServiceSpec{
				Ports: []v1.ServicePort{
					{
						Name:       "http-function-port",
						Port:       fPort,
						TargetPort: intstr.FromInt(int(fPort)),
						NodePort:   0,
						Protocol:   v1.ProtocolTCP,
					},
				},
				Type: v1.ServiceTypeClusterIP,
			},
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
									},
								},
							},
						},
					},
				},
			},
		},
	}
	return &f
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

	langruntime.AddFakeConfig(clientset)
	lr := langruntime.SetupLangRuntime(clientset)
	lr.ReadConfigMap()

	f1Name := "f1"
	f1 := getDefaultFunc(f1Name, ns)
	f1Port := f1.Spec.ServiceSpec.Ports[0].Port
	f1.ObjectMeta.Labels = funcLabels
	f1.Spec.Deployment.ObjectMeta = metav1.ObjectMeta{
		Annotations: funcAnno,
	}
	f1.Spec.Deployment.Spec.Template.ObjectMeta = metav1.ObjectMeta{
		Annotations: funcAnno,
	}
	// Testing happy path
	pullSecrets := []v1.LocalObjectReference{
		{Name: "creds"},
	}
	err := EnsureFuncDeployment(clientset, f1, or, lr, "", "unzip", pullSecrets)
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
		Annotations:     funcAnno,
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
				Name:  "FUNC_RUNTIME",
				Value: "python2.7",
			},
			{
				Name:  "FUNC_MEMORY_LIMIT",
				Value: "0",
			},
			{
				Name:  "FUNC_PORT",
				Value: strconv.Itoa(int(f1Port)),
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

	secrets := dpm.Spec.Template.Spec.ImagePullSecrets
	if secrets[0].Name != "creds" && secrets[1].Name != "p1" && secrets[2].Name != "p2" {
		t.Errorf("Expected first secret to be 'p1' but found %v and second secret to be 'p2' and found %v", secrets[0], secrets[1])
	}

	// Init containers behavior should be tested with integration tests
	if len(dpm.Spec.Template.Spec.InitContainers) < 1 {
		t.Errorf("Expecting at least an init container to install deps")
	}
	if dpm.Spec.Template.Spec.InitContainers[0].Image != "unzip" {
		t.Errorf("Unexpected init image %s", dpm.Spec.Template.Spec.InitContainers[0].Image)
	}

	// If no handler and function is given it should not fail
	f2 := getDefaultFunc("func2", ns)
	f2.Spec.Function = ""
	f2.Spec.Handler = ""
	err = EnsureFuncDeployment(clientset, f2, or, lr, "", "unzip", []v1.LocalObjectReference{})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	dpm, err = clientset.ExtensionsV1beta1().Deployments(ns).Get("func2", metav1.GetOptions{})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	// If the Image has been already provided it should not resolve it
	f3 := getDefaultFunc("func3", ns)
	f3.Spec.Deployment.Spec.Template.Spec.Containers[0].Image = "test-image"
	err = EnsureFuncDeployment(clientset, f3, or, lr, "", "unzip", []v1.LocalObjectReference{})
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
	f4 := getDefaultFunc("func4", ns)
	f4.Spec.Function = ""
	f4.Spec.Deps = ""
	err = EnsureFuncDeployment(clientset, f4, or, lr, "", "unzip", []v1.LocalObjectReference{})
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

	// It should update a deployment if it is already present
	f6 := kubelessApi.Function{}
	f6 = *f1
	f6.Spec.Handler = "foo.bar2"
	f6.Spec.Deployment.ObjectMeta.Annotations["new-key"] = "value"
	err = EnsureFuncDeployment(clientset, &f6, or, lr, "", "unzip", []v1.LocalObjectReference{})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	// Unable to ensure that the new deployment is patched since fake
	// ignores PATCH actions: https://github.com/kubernetes/client-go/issues/364

	// It should return an error if some dependencies are given but the runtime is not supported
	f7 := getDefaultFunc("func7", ns)
	f7.Spec.Deps = "deps"
	f7.Spec.Runtime = "cobol"
	err = EnsureFuncDeployment(clientset, f7, or, lr, "", "unzip", []v1.LocalObjectReference{})

	if err == nil {
		t.Fatal("An error should be thrown")
	}

	// If a timeout is specified it should set an environment variable FUNC_TIMEOUT
	f8 := getDefaultFunc("func8", ns)
	f8.Spec.Timeout = "10"
	err = EnsureFuncDeployment(clientset, f8, or, lr, "", "unzip", []v1.LocalObjectReference{})
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

	// If a prebuilt image is specified it should not build the function using init containers
	f9 := getDefaultFunc("func9", ns)
	err = EnsureFuncDeployment(clientset, f9, or, lr, "user/image:test", "unzip", []v1.LocalObjectReference{})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	dpm, err = clientset.ExtensionsV1beta1().Deployments(ns).Get("func9", metav1.GetOptions{})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if dpm.Spec.Template.Spec.Containers[0].Image != "user/image:test" {
		t.Errorf("Unexpected image %s, expecting prebuilt user/image:test", dpm.Spec.Template.Spec.Containers[0].Image)
	}
	if len(dpm.Spec.Template.Spec.InitContainers) != 0 {
		t.Error("Unexpected init containers")
	}

	// It should include existing volumes
	f10 := getDefaultFunc("func10", ns)
	f10.Spec.Deployment.Spec.Template.Spec.Volumes = []v1.Volume{
		{
			Name:         "test",
			VolumeSource: v1.VolumeSource{},
		},
	}
	f10.Spec.Deployment.Spec.Template.Spec.Containers[0].VolumeMounts = []v1.VolumeMount{
		{
			Name:      "test",
			MountPath: "/tmp/test",
		},
	}
	err = EnsureFuncDeployment(clientset, f10, or, lr, "", "unzip", []v1.LocalObjectReference{})
	dpm, err = clientset.ExtensionsV1beta1().Deployments(ns).Get("func10", metav1.GetOptions{})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if dpm.Spec.Template.Spec.Volumes[0].Name != "test" {
		t.Error("Should maintain volumen test")
	}
	if dpm.Spec.Template.Spec.Containers[0].VolumeMounts[0].Name != "test" {
		t.Error("Should maintain volumen test")
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
		NegotiatedSerializer: scheme.Codecs,
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
			APIVersion: "kubeless.io/v1beta1",
		},
	}
	ns := "default"
	f1Name := "func1"
	f1 := &kubelessApi.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name:      f1Name,
			Namespace: ns,
		},
		Spec: kubelessApi.FunctionSpec{
			Timeout: "120",
		},
	}
	expectedMeta := metav1.ObjectMeta{
		Name:            "trigger-" + f1Name,
		Namespace:       ns,
		OwnerReferences: or,
	}

	clientset := fake.NewSimpleClientset()

	pullSecrets := []v1.LocalObjectReference{
		{Name: "creds"},
	}
	err := EnsureCronJob(clientset, f1, "* * * * *", "unzip", or, pullSecrets)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	cronJob, err := clientset.BatchV1beta1().CronJobs(ns).Get(fmt.Sprintf("trigger-%s", f1.Name), metav1.GetOptions{})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
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
	expectedCommand := []string{"curl", "-Lv", fmt.Sprintf("http://%s.%s.svc.cluster.local:8080", f1Name, ns)}
	runtimeContainer := cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0]
	if runtimeContainer.Image != "unzip" {
		t.Errorf("Unexpected image %s", runtimeContainer.Image)
	}
	args := runtimeContainer.Args
	// skip event headers data (i.e  -H "event-id: cronjob-controller-2018-03-05T05:55:41.990784027Z" etc)
	foundCommand := []string{args[0], args[1], args[len(args)-1]}
	if !reflect.DeepEqual(foundCommand, expectedCommand) {
		t.Errorf("Unexpected command %s expexted %s", foundCommand, expectedCommand)
	}

	// It should update the existing cronJob if it is already created
	newSchedule := "*/10 * * * *"
	err = EnsureCronJob(clientset, f1, newSchedule, "unzip", or, pullSecrets)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	updatedCronJob, err := clientset.BatchV1beta1().CronJobs(ns).Get(fmt.Sprintf("trigger-%s", f1.Name), metav1.GetOptions{})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if updatedCronJob.Spec.Schedule != newSchedule {
		t.Errorf("Unexpected schedule %s expecting %s", updatedCronJob.Spec.Schedule, newSchedule)
	}
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
	fakeSvc := v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "myns",
			Name:      "foo",
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{TargetPort: intstr.FromInt(8080)},
			},
		},
	}
	clientset := fake.NewSimpleClientset(&fakeSvc)
	f1 := &kubelessApi.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "myns",
			UID:       "1234",
		},
		Spec: kubelessApi.FunctionSpec{},
	}
	httpTrigger := &kubelessApi.HTTPTrigger{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "myns",
			UID:       "1234",
		},
		Spec: kubelessApi.HTTPTriggerSpec{
			FunctionName: f1.Name,
		},
	}
	if err := CreateIngress(clientset, httpTrigger); err != nil {
		t.Fatalf("Creating ingress returned err: %v", err)
	}
	if err := CreateIngress(clientset, httpTrigger); err != nil {
		if !k8sErrors.IsAlreadyExists(err) {
			t.Fatalf("Expect object is already exists, got %v", err)
		}
	}
}

func TestCreateIngressResourceWithNginxGateway(t *testing.T) {
	fakeSvc := v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "myns",
			Name:      "foo",
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{TargetPort: intstr.FromInt(8080)},
			},
		},
	}
	clientset := fake.NewSimpleClientset(&fakeSvc)
	f1 := &kubelessApi.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "myns",
			UID:       "1234",
		},
		Spec: kubelessApi.FunctionSpec{},
	}
	httpTrigger := &kubelessApi.HTTPTrigger{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "myns",
			UID:       "1234",
		},
		Spec: kubelessApi.HTTPTriggerSpec{
			HostName:        "foo",
			TLSAcme:         true,
			FunctionName:    f1.Name,
			BasicAuthSecret: "foo-secret",
			Gateway:         "nginx",
		},
	}
	if err := CreateIngress(clientset, httpTrigger); err != nil {
		t.Fatalf("Creating ingress returned err: %v", err)
	}

	ingress, err := clientset.ExtensionsV1beta1().Ingresses("myns").Get("foo", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Getting Ingress returned err: %v", err)
	}

	annotations := ingress.ObjectMeta.Annotations
	if annotations == nil || len(annotations) == 0 ||
		annotations["kubernetes.io/ingress.class"] != "nginx" ||
		annotations["nginx.ingress.kubernetes.io/auth-secret"] != "foo-secret" ||
		annotations["nginx.ingress.kubernetes.io/auth-type"] != "basic" ||
		annotations["kubernetes.io/tls-acme"] != "true" ||
		annotations["nginx.ingress.kubernetes.io/ssl-redirect"] != "true" {
		t.Fatal("Missing or wrong annotations!")
	}

	tls := ingress.Spec.TLS
	if tls == nil || len(tls) != 1 ||
		tls[0].SecretName == "" ||
		tls[0].Hosts == nil || len(tls[0].Hosts) != 1 || tls[0].Hosts[0] == "" {
		t.Fatal("Missing or incomplete TLS spec!")
	}
}

func TestCreateIngressResourceWithTLSAcme(t *testing.T) {
	fakeSvc := v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "myns",
			Name:      "foo",
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{TargetPort: intstr.FromInt(8080)},
			},
		},
	}
	clientset := fake.NewSimpleClientset(&fakeSvc)
	f1 := &kubelessApi.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "myns",
			UID:       "1234",
		},
		Spec: kubelessApi.FunctionSpec{},
	}
	httpTrigger := &kubelessApi.HTTPTrigger{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "myns",
			UID:       "1234",
		},
		Spec: kubelessApi.HTTPTriggerSpec{
			HostName:     "foo",
			TLSAcme:      true,
			FunctionName: f1.Name,
		},
	}
	if err := CreateIngress(clientset, httpTrigger); err != nil {
		t.Fatalf("Creating ingress returned err: %v", err)
	}

	ingress, err := clientset.ExtensionsV1beta1().Ingresses("myns").Get("foo", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Getting Ingress returned err: %v", err)
	}

	annotations := ingress.ObjectMeta.Annotations
	if annotations == nil || len(annotations) == 0 ||
		annotations["kubernetes.io/tls-acme"] != "true" ||
		annotations["nginx.ingress.kubernetes.io/ssl-redirect"] != "true" {
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
			NegotiatedSerializer: scheme.Codecs,
		},
	}
}

func TestUpdateIngressResource(t *testing.T) {
	fakeSvc := v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "myns",
			Name:      "foo",
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{TargetPort: intstr.FromInt(8080)},
			},
		},
	}
	fakeIngress := v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "myns",
			Name:      "foo",
			Labels: map[string]string{
				"test": "foo",
			},
		},
	}
	clientset := fake.NewSimpleClientset(&fakeSvc, &fakeIngress)
	f1 := &kubelessApi.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "myns",
			UID:       "1234",
		},
		Spec: kubelessApi.FunctionSpec{},
	}
	httpTrigger := &kubelessApi.HTTPTrigger{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "myns",
			UID:       "1234",
		},
		Spec: kubelessApi.HTTPTriggerSpec{
			FunctionName: f1.Name,
			HostName:     "test.domain",
		},
	}
	if err := CreateIngress(clientset, httpTrigger); err != nil {
		t.Fatalf("Creating ingress returned err: %v", err)
	}
	newIngress, err := clientset.ExtensionsV1beta1().Ingresses("myns").Get("foo", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Unexpected errors: %v", err)
	}
	if newIngress.ObjectMeta.Labels["test"] != "foo" {
		t.Errorf("Unexpected labels: %v", newIngress.ObjectMeta.Labels)
	}
	if newIngress.Spec.Rules[0].Host != "test.domain" {
		t.Errorf("Unexpected hostname: %s", newIngress.Spec.Rules[0].Host)
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

func TestGetProvisionContainer(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	langruntime.AddFakeConfig(clientset)
	lr := langruntime.SetupLangRuntime(clientset)
	lr.ReadConfigMap()

	rvol := v1.VolumeMount{Name: "runtime", MountPath: "/runtime"}
	dvol := v1.VolumeMount{Name: "deps", MountPath: "/deps"}
	c, err := getProvisionContainer("test", "sha256:abc1234", "test.func", "test.foo", "text", "python2.7", "unzip", rvol, dvol, lr)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	expectedContainer := v1.Container{
		Name:            "prepare",
		Image:           "unzip",
		Command:         []string{"sh", "-c"},
		Args:            []string{"echo 'abc1234  /deps/test.func' > /tmp/func.sha256 && sha256sum -c /tmp/func.sha256 && cp /deps/test.func /runtime/test.py && cp /deps/requirements.txt /runtime"},
		VolumeMounts:    []v1.VolumeMount{rvol, dvol},
		ImagePullPolicy: v1.PullIfNotPresent,
	}
	if !reflect.DeepEqual(expectedContainer, c) {
		t.Errorf("Unexpected result:\n %+v", c)
	}

	// If the content type is encoded it should decode it
	c, err = getProvisionContainer("Zm9vYmFyCg==", "sha256:abc1234", "test.func", "test.foo", "base64", "python2.7", "unzip", rvol, dvol, lr)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if !strings.HasPrefix(c.Args[0], "base64 -d < /deps/test.func > /tmp/func.decoded") {
		t.Errorf("Unexpected command: %s", c.Args[0])
	}

	secrets, err := lr.GetImageSecrets("python2.7")
	if err != nil {
		t.Errorf("Unable to fetch secrets: %v", err)
	}

	if secrets[0].Name != "p1" && secrets[1].Name != "p2" {
		t.Errorf("Expected first secret to be 'p1' but found %v and second secret to be 'p2' but found %v", secrets[0], secrets[1])
	}

	// It should skip the dependencies installation if the runtime is not supported
	c, err = getProvisionContainer("function", "sha256:abc1234", "test.func", "test.foo", "text", "cobol", "unzip", rvol, dvol, lr)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if strings.Contains(c.Args[0], "cp /deps ") {
		t.Errorf("Unexpected command: %s", c.Args[0])
	}

	// It should extract the file in case it is a Zip
	c, err = getProvisionContainer("Zm9vYmFyCg==", "sha256:abc1234", "test.zip", "test.foo", "base64+zip", "python2.7", "unzip", rvol, dvol, lr)
	if !strings.Contains(c.Args[0], "unzip -o /tmp/func.decoded -d /runtime") {
		t.Errorf("Unexpected command: %s", c.Args[0])
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
