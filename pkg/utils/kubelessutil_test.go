package utils

import (
	"reflect"
	"strconv"
	"strings"
	"testing"

	kubelessApi "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	"github.com/kubeless/kubeless/pkg/langruntime"

	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/fake"
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
}

func TestEnsureFuncMapWithoutDeps(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	or := []metav1.OwnerReference{
		{
			Kind:       "Function",
			APIVersion: "kubeless.io/v1beta1",
		},
	}
	ns := "default"
	langruntime.AddFakeConfig(clientset)
	lr := langruntime.SetupLangRuntime(clientset)
	lr.ReadConfigMap()
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

	err := EnsureFuncConfigMap(clientset, f2, or, lr)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	cm, err := clientset.CoreV1().ConfigMaps(ns).Get(f2Name, metav1.GetOptions{})
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

func TestAvoidConfigMapOverwrite(t *testing.T) {
	f1Name := "f1"
	clientset, or, ns, lr := prepareDeploymentTest(f1Name)
	clientset.CoreV1().ConfigMaps(ns).Create(&v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      f1Name,
			Namespace: ns,
		},
	})
	f1 := getDefaultFunc(f1Name, ns)
	err := EnsureFuncConfigMap(clientset, f1, or, lr)
	if err == nil && strings.Contains(err.Error(), "conflicting object") {
		t.Errorf("It should fail because a conflict")
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
}

func TestUpdateFuncSvc(t *testing.T) {
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
	// If there is already a service it should update the previous one
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
	newLabels := map[string]string{
		"foobar": "barfoo",
	}
	f1.ObjectMeta.Labels = newLabels
	err = EnsureFuncService(clientset, f1, or)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	svc, err := clientset.CoreV1().Services(ns).Get(f1Name, metav1.GetOptions{})
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

func TestAvoidServiceOverwrite(t *testing.T) {
	f1Name := "f1"
	ns := "default"
	or := []metav1.OwnerReference{
		{
			Kind:       "Function",
			APIVersion: "kubeless.io/v1beta1",
		},
	}
	clientset := fake.NewSimpleClientset()
	clientset.CoreV1().Services(ns).Create(&v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      f1Name,
			Namespace: ns,
		},
	})
	f1 := getDefaultFunc(f1Name, ns)
	err := EnsureFuncService(clientset, f1, or)
	if err == nil && strings.Contains(err.Error(), "conflicting object") {
		t.Errorf("It should fail because a conflict")
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

func prepareDeploymentTest(funcName string) (*fake.Clientset, []metav1.OwnerReference, string, *langruntime.Langruntimes) {
	clientset := fake.NewSimpleClientset()
	or := []metav1.OwnerReference{
		{
			Kind:       "Function",
			APIVersion: "k8s.io",
		},
	}
	ns := "default"
	langruntime.AddFakeConfig(clientset)
	lr := langruntime.SetupLangRuntime(clientset)
	lr.ReadConfigMap()
	return clientset, or, ns, lr
}

func TestEnsureDeployment(t *testing.T) {
	f1Name := "f1"
	clientset, or, ns, lr := prepareDeploymentTest(f1Name)
	funcLabels := map[string]string{
		"foo": "bar",
	}
	funcAnno := map[string]string{
		"bar": "foo",
	}
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
		Labels:          addDefaultLabel(funcLabels),
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
		Image: "bar",
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
				Name:  "KUBELESS_INSTALL_VOLUME",
				Value: "/kubeless",
			},
			{
				Name:  "PYTHONPATH",
				Value: "/kubeless/lib/python2.7/site-packages:/kubeless",
			},
		},
		VolumeMounts: []v1.VolumeMount{
			{
				Name:      f1Name,
				MountPath: "/kubeless",
			},
		},
		LivenessProbe: &v1.Probe{
			InitialDelaySeconds: int32(5),
			PeriodSeconds:       int32(10),
			Handler: v1.Handler{
				Exec: &v1.ExecAction{
					Command: []string{"curl", "-f", "http://localhost:8080/healthz"},
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
}

func TestEnsureDeploymentWithoutFuncNorHandler(t *testing.T) {
	funcName := "func2"
	clientset, or, ns, lr := prepareDeploymentTest(funcName)
	// If no handler and function is given it should not fail
	f2 := getDefaultFunc(funcName, ns)
	f2.Spec.Function = ""
	f2.Spec.Handler = ""
	err := EnsureFuncDeployment(clientset, f2, or, lr, "", "unzip", []v1.LocalObjectReference{})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	_, err = clientset.ExtensionsV1beta1().Deployments(ns).Get(funcName, metav1.GetOptions{})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
}

func TestEnsureDeploymentWithImage(t *testing.T) {
	funcName := "func"
	clientset, or, ns, lr := prepareDeploymentTest(funcName)
	// If the Image has been already provided it should not resolve it
	f3 := getDefaultFunc(funcName, ns)
	f3.Spec.Deployment.Spec.Template.Spec.Containers[0].Image = "test-image"
	err := EnsureFuncDeployment(clientset, f3, or, lr, "", "unzip", []v1.LocalObjectReference{})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	dpm, err := clientset.ExtensionsV1beta1().Deployments(ns).Get(funcName, metav1.GetOptions{})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if dpm.Spec.Template.Spec.Containers[0].Image != "test-image" {
		t.Errorf("Unexpected Image Name: %s", dpm.Spec.Template.Spec.Containers[0].Image)
	}
}

func TestEnsureDeploymentWithoutFunc(t *testing.T) {
	funcName := "func"
	clientset, or, ns, lr := prepareDeploymentTest(funcName)
	// If no function is given it should not use an init container
	f4 := getDefaultFunc(funcName, ns)
	f4.Spec.Function = ""
	f4.Spec.Deps = ""
	err := EnsureFuncDeployment(clientset, f4, or, lr, "", "unzip", []v1.LocalObjectReference{})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	dpm, err := clientset.ExtensionsV1beta1().Deployments(ns).Get(funcName, metav1.GetOptions{})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if len(dpm.Spec.Template.Spec.InitContainers) > 0 {
		t.Error("It should not setup an init container")
	}
}

func TestEnsureUpdateDeployment(t *testing.T) {
	f1Name := "f1"
	clientset, or, ns, lr := prepareDeploymentTest(f1Name)
	// It should update a deployment if it is already present
	funcAnno := map[string]string{
		"bar": "foo",
	}
	f1 := getDefaultFunc(f1Name, ns)
	f1.Spec.Deployment.ObjectMeta = metav1.ObjectMeta{
		Annotations: funcAnno,
	}
	f1.Spec.Deployment.Spec.Template.ObjectMeta = metav1.ObjectMeta{
		Annotations: funcAnno,
	}
	err := EnsureFuncDeployment(clientset, f1, or, lr, "", "unzip", []v1.LocalObjectReference{})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
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
}

func TestAvoidDeploymentOverwrite(t *testing.T) {
	f1Name := "f1"
	clientset, or, ns, lr := prepareDeploymentTest(f1Name)
	clientset.ExtensionsV1beta1().Deployments(ns).Create(&v1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      f1Name,
			Namespace: ns,
		},
	})
	f1 := getDefaultFunc(f1Name, ns)
	err := EnsureFuncDeployment(clientset, f1, or, lr, "", "unzip", []v1.LocalObjectReference{})
	if err == nil && strings.Contains(err.Error(), "conflicting object") {
		t.Errorf("It should fail because a conflict")
	}
}

func TestDeploymentWithUnsupportedRuntime(t *testing.T) {
	funcName := "func"
	clientset, or, ns, lr := prepareDeploymentTest(funcName)
	// It should return an error if some dependencies are given but the runtime is not supported
	f7 := getDefaultFunc("func7", ns)
	f7.Spec.Deps = "deps"
	f7.Spec.Runtime = "cobol"
	err := EnsureFuncDeployment(clientset, f7, or, lr, "", "unzip", []v1.LocalObjectReference{})

	if err == nil {
		t.Fatal("An error should be thrown")
	}
}

func TestDeploymentWithTimeout(t *testing.T) {
	funcName := "func"
	clientset, or, ns, lr := prepareDeploymentTest(funcName)
	// If a timeout is specified it should set an environment variable FUNC_TIMEOUT
	f8 := getDefaultFunc(funcName, ns)
	f8.Spec.Timeout = "10"
	err := EnsureFuncDeployment(clientset, f8, or, lr, "", "unzip", []v1.LocalObjectReference{})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	dpm, err := clientset.ExtensionsV1beta1().Deployments(ns).Get(funcName, metav1.GetOptions{})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if getEnvValueFromList("FUNC_TIMEOUT", dpm.Spec.Template.Spec.Containers[0].Env) != "10" {
		t.Error("Unable to set timeout")
	}
}

func TestDeploymentWithPrebuiltImage(t *testing.T) {
	funcName := "func"
	clientset, or, ns, lr := prepareDeploymentTest(funcName)
	// If a prebuilt image is specified it should not build the function using init containers
	f9 := getDefaultFunc(funcName, ns)
	err := EnsureFuncDeployment(clientset, f9, or, lr, "user/image:test", "unzip", []v1.LocalObjectReference{})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	dpm, err := clientset.ExtensionsV1beta1().Deployments(ns).Get(funcName, metav1.GetOptions{})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if dpm.Spec.Template.Spec.Containers[0].Image != "user/image:test" {
		t.Errorf("Unexpected image %s, expecting prebuilt user/image:test", dpm.Spec.Template.Spec.Containers[0].Image)
	}
	if len(dpm.Spec.Template.Spec.InitContainers) != 0 {
		t.Error("Unexpected init containers")
	}
}

func TestDeploymentWithVolumes(t *testing.T) {
	funcName := "func"
	clientset, or, ns, lr := prepareDeploymentTest(funcName)
	// It should include existing volumes
	f10 := getDefaultFunc(funcName, ns)
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
	err := EnsureFuncDeployment(clientset, f10, or, lr, "", "unzip", []v1.LocalObjectReference{})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	dpm, err := clientset.ExtensionsV1beta1().Deployments(ns).Get(funcName, metav1.GetOptions{})
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

func doesNotContain(envs []v1.EnvVar, env v1.EnvVar) bool {
	for _, e := range envs {
		if e == env {
			return false
		}
	}
	return true
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

	// If the content type is url it should use curl
	c, err = getProvisionContainer("https://raw.githubusercontent.com/test/test/test/test.py", "sha256:abc1234", "", "test.foo", "url", "python2.7", "unzip", rvol, dvol, lr)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if !strings.HasPrefix(c.Args[0], "curl https://raw.githubusercontent.com/test/test/test/test.py -L --silent --output /tmp/func.fromurl") {
		t.Errorf("Unexpected command: %s", c.Args[0])
	}

	// If the content type is url it should use curl
	c, err = getProvisionContainer("https://raw.githubusercontent.com/test/test/test/test.py", "sha256:abc1234", "", "test.foo", "url+zip", "python2.7", "unzip", rvol, dvol, lr)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if !strings.HasPrefix(c.Args[0], "curl https://raw.githubusercontent.com/test/test/test/test.py -L --silent --output /tmp/func.fromurl") {
		t.Errorf("Unexpected command: %s", c.Args[0])
	}
}
