package utils

import (
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/kubeless/kubeless/pkg/spec"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
	av2alpha1 "k8s.io/client-go/pkg/apis/autoscaling/v2alpha1"
	"k8s.io/client-go/pkg/apis/batch/v2alpha1"
	xv1beta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/rest"
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

func TestDeleteK8sResources(t *testing.T) {
	myNsFoo := metav1.ObjectMeta{
		Namespace: "myns",
		Name:      "foo",
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

	clientset := fake.NewSimpleClientset(&deploy, &svc, &cm)

	if err := DeleteK8sResources("myns", "foo", clientset); err != nil {
		t.Fatalf("Deleting resources returned err: %v", err)
	}

	t.Log("Actions:", clientset.Actions())

	for _, kind := range []string{"services", "configmaps", "deployments"} {
		a := findAction(clientset, "delete", kind)
		if a == nil {
			t.Errorf("failed to delete %s", kind)
		} else if ns := a.GetNamespace(); ns != "myns" {
			t.Errorf("deleted %s from wrong namespace (%s)", kind, ns)
		} else if n := a.(ktesting.DeleteAction).GetName(); n != "foo" {
			t.Errorf("deleted %s with wrong name (%s)", kind, n)
		}
	}

	// Similar with only svc remaining
	clientset = fake.NewSimpleClientset(&svc)

	if err := DeleteK8sResources("myns", "foo", clientset); err != nil {
		t.Fatalf("Deleting partial resources returned err: %v", err)
	}

	t.Log("Actions:", clientset.Actions())

	if !hasAction(clientset, "delete", "services") {
		t.Errorf("failed to delete service")
	}

	// Test deleting cronjob
	job := v2alpha1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "myns",
			Name:      "trigger-foo",
		},
	}

	clientset = fake.NewSimpleClientset(&job, &deploy, &svc, &cm)

	if err := DeleteK8sResources("myns", "foo", clientset); err != nil {
		t.Fatalf("Deleting resources returned err: %v", err)
	}

	t.Log("Actions:", clientset.Actions())

	for _, kind := range []string{"cronjobs", "services", "configmaps", "deployments"} {
		a := findAction(clientset, "delete", kind)
		if a == nil {
			t.Errorf("failed to delete %s", kind)
		} else if ns := a.GetNamespace(); ns != "myns" {
			t.Errorf("deleted %s from wrong namespace (%s)", kind, ns)
		}
	}
}
func TestEnsureK8sResources(t *testing.T) {

	myNsFoo := metav1.ObjectMeta{
		Namespace: "kubeless",
		Name:      "minio-key",
	}

	secret := v1.Secret{
		ObjectMeta: myNsFoo,
	}

	clientset := fake.NewSimpleClientset(&secret)
	ns := "myns"
	func1 := "foo1"

	funcLabels := map[string]string{
		"foo": "bar",
	}
	funcAnno := map[string]string{
		"bar": "foo",
	}

	f1 := &spec.Function{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Function",
			APIVersion: "k8s.io/v1",
		},
		Metadata: metav1.ObjectMeta{
			Name:      func1,
			Namespace: ns,
			Labels:    funcLabels,
		},
		Spec: spec.FunctionSpec{
			Handler: "foo.bar",
			Runtime: "python2.7",
		},
	}

	if err := EnsureK8sResources(f1, clientset); err != nil {
		t.Fatalf("Creating resources returned err: %v", err)
	}

	svc, err := clientset.CoreV1().Services(ns).Get(func1, metav1.GetOptions{})
	if err != nil {
		t.Errorf("Can't create svc object: %v", err)
	}
	if svc.ObjectMeta.Name != func1 {
		t.Errorf("Create wrong svc object. Expect svc name is %s but got %s", func1, svc.ObjectMeta.Name)
	}

	cm, err := clientset.CoreV1().ConfigMaps(ns).Get(func1, metav1.GetOptions{})
	if err != nil {
		t.Errorf("Can't create configmap object: %v", err)
	}
	if cm.Data["handler"] != "foo.bar" {
		t.Errorf("Create wrong configmap object. Expect configmap data handler is foo.bar but got %s", cm.Data["handler"])
	}

	dpm, err := clientset.ExtensionsV1beta1().Deployments(ns).Get(func1, metav1.GetOptions{})
	if err != nil {
		t.Errorf("Can't create deployment object: %v", err)
	}
	if dpm.Spec.Template.Labels["foo"] != "bar" {
		t.Errorf("Create wrong deployment object. Expect deployment labels foo=bar but got %s", dpm.Spec.Template.Labels["foo"])
	}

	func2 := "foo2"
	f2 := &spec.Function{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Function",
			APIVersion: "k8s.io/v1",
		},
		Metadata: metav1.ObjectMeta{
			Name:      func2,
			Namespace: ns,
			Labels:    funcLabels,
		},
		Spec: spec.FunctionSpec{
			Handler: "foo.bar",
			Runtime: "python2.7",
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
	if err := EnsureK8sResources(f2, clientset); err != nil {
		t.Fatalf("Creating resources returned err: %v", err)
	}
	svc, err = clientset.CoreV1().Services(ns).Get(func2, metav1.GetOptions{})
	if err != nil {
		t.Errorf("Can't create svc object: %v", err)
	}
	if svc.ObjectMeta.Name != func2 {
		t.Errorf("Create wrong svc object. Expect svc name is %s but got %s", func2, svc.ObjectMeta.Name)
	}

	cm, err = clientset.CoreV1().ConfigMaps(ns).Get(func2, metav1.GetOptions{})
	if err != nil {
		t.Errorf("Can't create configmap object: %v", err)
	}

	dpm, err = clientset.ExtensionsV1beta1().Deployments(ns).Get(func2, metav1.GetOptions{})
	if err != nil {
		t.Errorf("Can't create deployment object: %v", err)
	}
	if len(dpm.Spec.Template.Spec.Containers[0].Env) == 0 {
		t.Errorf("There is no environment variable in the deployment")
	}
	if doesNotContain(dpm.Spec.Template.Spec.Containers[0].Env, v1.EnvVar{
		Name:  "foo",
		Value: "bar",
	}) {
		t.Errorf("Deployment env doesn't contain foo=bar")
	}
	if v, ok := dpm.Spec.Template.ObjectMeta.Annotations["bar"]; ok {
		if v != "foo" {
			t.Errorf("Deployment annotation doesn't contain bar=foo")
		}
	} else {
		t.Errorf("Deployment annotation doesn't contain key bar")
	}

	func3 := "foo3"
	f3 := &spec.Function{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Function",
			APIVersion: "k8s.io/v1",
		},
		Metadata: metav1.ObjectMeta{
			Name:      func3,
			Namespace: ns,
			Labels:    funcLabels,
		},
		Spec: spec.FunctionSpec{
			Handler: "foo.bar",
			Runtime: "nodejs6",
			Deps:    "sample",
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Env: []v1.EnvVar{
								{
									Name:  "NPM_REGISTRY",
									Value: "http://my-registry.com",
								},
								{
									Name:  "NPM_SCOPE",
									Value: "@kubeless",
								},
							},
						},
					},
				},
			},
		},
	}
	if err := EnsureK8sResources(f3, clientset); err != nil {
		t.Fatalf("Creating resources returned err: %v", err)
	}
	dpm, err = clientset.ExtensionsV1beta1().Deployments(ns).Get(func3, metav1.GetOptions{})
	t.Log(dpm.Spec.Template.Spec.InitContainers[1].Args)
	if len(dpm.Spec.Template.Spec.InitContainers) != 2 {
		t.Errorf("Init container should be defined when dependencies are specified")
	}
	if match, _ := regexp.MatchString("npm config set @kubeless:registry http://my-registry.com", strings.Join(dpm.Spec.Template.Spec.InitContainers[1].Args, " ")); !match {
		t.Errorf("Unable to set the NPM registry")
	}

	func4 := "foo4"
	f4 := &spec.Function{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Function",
			APIVersion: "k8s.io/v1",
		},
		Metadata: metav1.ObjectMeta{
			Name:      func4,
			Namespace: ns,
			Labels:    funcLabels,
		},
		Spec: spec.FunctionSpec{
			Handler: "foo.bar",
			Runtime: "python2.6",
			Deps:    "sample",
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Image: "custom-image",
						},
					},
				},
			},
		},
	}
	if err := EnsureK8sResources(f4, clientset); err != nil {
		t.Fatalf("Creating resources returned err: %v", err)
	}
	dpm, err = clientset.ExtensionsV1beta1().Deployments(ns).Get(func4, metav1.GetOptions{})
	if len(dpm.Spec.Template.Spec.InitContainers) != 2 {
		t.Errorf("Init container should be defined when dependencies are specified")
	}
	match, err := regexp.MatchString("custom-image", dpm.Spec.Template.Spec.Containers[0].Image)
	if err != nil {
		t.Error(err)
	}
	if !match {
		t.Errorf("Unable to set a custom image ID")
	}

	func5 := "foo5"
	f5 := &spec.Function{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Function",
			APIVersion: "k8s.io/v1",
		},
		Metadata: metav1.ObjectMeta{
			Name:      func5,
			Namespace: ns,
			Labels:    funcLabels,
		},
		Spec: spec.FunctionSpec{
			Handler: "foo.bar",
			Runtime: "csharp",
			Deps:    "sample",
		},
	}
	if err := EnsureK8sResources(f5, clientset); err == nil {
		t.Fatalf("Creating resources for an unsupported runtime with dependencies should return an error")
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
	clientset := fake.NewSimpleClientset()
	if err := CreateIngress(clientset, "foo", "bar", "foo.bar", "myns", false); err != nil {
		t.Fatalf("Creating ingress returned err: %v", err)
	}
	if err := CreateIngress(clientset, "foo", "bar", "foo.bar", "myns", false); err != nil {
		if !k8sErrors.IsAlreadyExists(err) {
			t.Fatalf("Expect object is already exists, got %v", err)
		}
	}
}

func TestCreateIngressResourceWithTLSAcme(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	if err := CreateIngress(clientset, "foo", "bar", "foo.bar", "myns", true); err != nil {
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
			GroupVersion:         &api.Registry.GroupOrDie(api.GroupName).GroupVersion,
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
	min := int32(1)
	max := int32(10)
	clientset := fake.NewSimpleClientset()
	if err := CreateAutoscale(clientset, "foo", "myns", "cpu", min, max, "50"); err != nil {
		t.Fatalf("Creating autoscale returned err: %v", err)
	}

	hpa, err := clientset.AutoscalingV2alpha1().HorizontalPodAutoscalers("myns").Get("foo", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Creating autoscale returned err: %v", err)
	}
	if hpa.Spec.ScaleTargetRef.Name != "foo" {
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
	c, err := getProvisionContainer("test", "sha256:abc1234", "test.func", "text", "test.foo", "python2.7", rvol, dvol)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	expectedContainer := v1.Container{
		Name:            "prepare",
		Image:           "kubeless/unzip@sha256:f162c062973cca05459834de6ed14c039d45df8cdb76097f50b028a1621b3697",
		Command:         []string{"sh", "-c"},
		Args:            []string{"cp /deps/test.func /runtime && cp /deps/requirements.txt /runtime && echo 'abc1234  /runtime/test.func' > /runtime/test.func.sha256 && sha256sum -c /runtime/test.func.sha256"},
		VolumeMounts:    []v1.VolumeMount{rvol, dvol},
		ImagePullPolicy: v1.PullIfNotPresent,
	}
	if !reflect.DeepEqual(expectedContainer, c) {
		t.Error("Unexpected result")
	}
}
