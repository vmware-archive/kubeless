package utils

import (
	"os"
	"testing"

	"github.com/kubeless/kubeless/pkg/spec"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/pkg/api/v1"
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
}

func check(runtime, ftype, fname string, values []string, t *testing.T) {
	imageName, depName, fileName, err := GetFunctionData(runtime, ftype, fname)
	if err != nil {
		t.Fatalf("Retrieving the image returned err: %v", err)
	}
	if imageName == "" {
		t.Fatalf("Retrieving the image returned an empty Image ID")
	}
	if depName != values[0] {
		t.Fatalf("Retrieving the image returned a wrong dependencies file. Received " + depName + " while expecting " + values[0])
	}
	if fileName != values[1] {
		t.Fatalf("Retrieving the image returned a wrong file name. Received " + fileName + " while expecting " + values[1])
	}
}
func TestGetFunctionData(t *testing.T) {

	expectedValues := []string{"requirements.txt", "test.py"}
	check("python2.7", "HTTP", "test", expectedValues, t)
	check("python2.7", "PubSub", "test", expectedValues, t)

	expectedValues = []string{"package.json", "test.js"}
	check("nodejs6", "HTTP", "test", expectedValues, t)
	check("nodejs6", "PubSub", "test", expectedValues, t)
	check("nodejs8", "HTTP", "test", expectedValues, t)
	check("nodejs8", "PubSub", "test", expectedValues, t)

	expectedValues = []string{"Gemfile", "test.rb"}
	check("ruby2.4", "HTTP", "test", expectedValues, t)

	_, _, _, err := GetFunctionData("unexistent", "HTTP", "test")
	if err == nil {
		t.Fatalf("Retrieving data for 'unexistent' should return an error")
	}

	expectedImageName := "ruby-test-image"
	os.Setenv("RUBY_RUNTIME", expectedImageName)
	imageR, _, _, errR := GetFunctionData("ruby", "HTTP", "test")
	if errR != nil {
		t.Fatalf("Retrieving the image returned err: %v", err)
	}
	if imageR != expectedImageName {
		t.Fatalf("Expecting " + imageR + " to be set to " + expectedImageName)
	}
	os.Unsetenv("RUBY_RUNTIME")

	expectedImageName = "ruby-pubsub-test-image"
	os.Setenv("RUBY_PUBSUB_RUNTIME", "ruby-pubsub-test-image")
	imageR, _, _, errR = GetFunctionData("ruby", "PubSub", "test")
	if errR != nil {
		t.Fatalf("Retrieving the image returned err: %v", err)
	}
	if imageR != expectedImageName {
		t.Fatalf("Expecting " + imageR + " to be set to " + expectedImageName)
	}
	os.Unsetenv("RUBY_PUBSUB_RUNTIME")
}

func TestEnsureK8sResources(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	ns := "myns"
	func1 := "foo1"
	func2 := "foo2"

	funcLabels := map[string]string{
		"foo": "bar",
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

	if err := EnsureK8sResources(ns, func1, f1, clientset); err != nil {
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

	if err := EnsureK8sResources(ns, func2, f2, clientset); err != nil {
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
}

func doesNotContain(envs []v1.EnvVar, env v1.EnvVar) bool {
	for _, e := range envs {
		if e == env {
			return false
		}
	}
	return true
}
