package controller

import (
	"reflect"
	"testing"

	"github.com/ghodss/yaml"
	kubelessApi "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	"github.com/kubeless/kubeless/pkg/langruntime"
	"github.com/sirupsen/logrus"
	"k8s.io/api/autoscaling/v2beta1"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	xv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	hpa := v2beta1.HorizontalPodAutoscaler{
		ObjectMeta: myNsFoo,
	}

	clientset := fake.NewSimpleClientset(&deploy, &svc, &cm, &hpa)

	controller := FunctionController{
		clientset: clientset,
	}
	if err := controller.deleteK8sResources("myns", "foo"); err != nil {
		t.Fatalf("Deleting resources returned err: %v", err)
	}

	t.Log("Actions:", clientset.Actions())

	for _, kind := range []string{"services", "configmaps", "deployments", "horizontalpodautoscalers"} {
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
	controller = FunctionController{
		clientset: clientset,
	}

	if err := controller.deleteK8sResources("myns", "foo"); err != nil {
		t.Fatalf("Deleting partial resources returned err: %v", err)
	}

	t.Log("Actions:", clientset.Actions())

	if !hasAction(clientset, "delete", "services") {
		t.Errorf("failed to delete service")
	}

	clientset = fake.NewSimpleClientset(&deploy, &svc, &cm)
	controller = FunctionController{
		clientset: clientset,
	}

	if err := controller.deleteK8sResources("myns", "foo"); err != nil {
		t.Fatalf("Deleting resources returned err: %v", err)
	}

	t.Log("Actions:", clientset.Actions())

	for _, kind := range []string{"services", "configmaps", "deployments"} {
		a := findAction(clientset, "delete", kind)
		if a == nil {
			t.Errorf("failed to delete %s", kind)
		} else if ns := a.GetNamespace(); ns != "myns" {
			t.Errorf("deleted %s from wrong namespace (%s)", kind, ns)
		}
	}
}

func TestEnsureK8sResourcesWithDeploymentDefinitionFromConfigMap(t *testing.T) {
	namespace := "default"
	funcName := "foo"
	var replicas int32
	replicas = 10
	funcLabels := map[string]string{
		"foo": "bar",
	}
	funcAnno := map[string]string{
		"bar": "foo",
		"xyz": "valuefromfunc",
	}
	funcObj := kubelessApi.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name:      funcName,
			Namespace: namespace,
			Labels:    funcLabels,
			UID:       "foo-uid",
		},
		Spec: kubelessApi.FunctionSpec{
			Function: "function",
			Deps:     "deps",
			Handler:  "foo.bar",
			Runtime:  "ruby2.4",
			Deployment: v1beta1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: funcAnno,
				},
				Spec: v1beta1.DeploymentSpec{
					Replicas: &replicas,
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
			},
		},
	}
	deploymentConfigData := `{
		"metadata": {
			"annotations": {
				"foo-from-deploy-cm": "bar-from-deploy-cm",
				"xyz": "valuefromcm"
			}
		},
		"spec": {
			"replicas": 2,
			"template": {
				"metadata": {
					"annotations": {
					"podannotation-from-func-crd": "value-from-container"
					}
				}
			}
		}
	}`

	runtimeImages := []langruntime.RuntimeInfo{{
		ID:             "ruby",
		DepName:        "Gemfile",
		FileNameSuffix: ".rb",
		Versions: []langruntime.RuntimeVersion{
			{
				Name:    "ruby24",
				Version: "2.4",
				Images: []langruntime.Image{
					{Phase: "runtime", Image: "bitnami/ruby:2.4"},
				},
				ImagePullSecrets: []langruntime.ImageSecret{},
			},
		},
	}}

	out, err := yaml.Marshal(runtimeImages)
	if err != nil {
		logrus.Fatal("Canot Marshall runtimeimage")
	}

	clientset := fake.NewSimpleClientset()
	kubelessConfigMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kubeless-config",
		},
		Data: map[string]string{"deployment": deploymentConfigData, "runtime-images": string(out)},
	}
	deploymentObjFromConfigMap := v1beta1.Deployment{}
	_ = yaml.Unmarshal([]byte(deploymentConfigData), &deploymentObjFromConfigMap)
	_, err = clientset.CoreV1().ConfigMaps(namespace).Create(kubelessConfigMap)
	if err != nil {
		logrus.Fatal("Unable to create configmap")
	}

	config, err := clientset.CoreV1().ConfigMaps(namespace).Get("kubeless-config", metav1.GetOptions{})
	if err != nil {
		logrus.Fatal("Unable to read the configmap")
	}
	var lr = langruntime.New(config)
	lr.ReadConfigMap()

	controller := FunctionController{
		logger:      logrus.WithField("pkg", "controller"),
		clientset:   clientset,
		langRuntime: lr,
		config:      config,
	}

	if err := controller.ensureK8sResources(&funcObj); err != nil {
		t.Fatalf("Creating/Updating resources returned err: %v", err)
	}
	dpm, _ := clientset.ExtensionsV1beta1().Deployments(namespace).Get(funcName, metav1.GetOptions{})
	expectedAnnotations := map[string]string{
		"bar":                "foo",
		"foo-from-deploy-cm": "bar-from-deploy-cm",
		"xyz":                "valuefromfunc",
	}
	for i := range expectedAnnotations {
		if dpm.ObjectMeta.Annotations[i] != expectedAnnotations[i] {
			t.Errorf("Expecting annotation %s but received %s", expectedAnnotations[i], dpm.ObjectMeta.Annotations[i])
		}
	}
	if *dpm.Spec.Replicas != 10 {
		t.Fatalf("Expecting replicas as 10 but received : %d", *dpm.Spec.Replicas)
	}
	expectedPodAnnotations := map[string]string{
		"bar":                "foo",
		"foo-from-deploy-cm": "bar-from-deploy-cm",
		"xyz":                "valuefromfunc",
		"podannotation-from-func-crd": "value-from-container",
	}
	for i := range expectedPodAnnotations {
		if dpm.Spec.Template.Annotations[i] != expectedPodAnnotations[i] {
			t.Fatalf("Expecting annotation %s but received %s", expectedPodAnnotations[i], dpm.ObjectMeta.Annotations[i])
		}
	}
}

func TestEnsureK8sResourcesWithLivenessProbeFromConfigMap(t *testing.T) {
	namespace := "default"
	funcName := "foo"
	var replicas int32
	replicas = 10
	funcLabels := map[string]string{
		"foo": "bar",
	}
	funcAnno := map[string]string{
		"bar": "foo",
		"xyz": "valuefromfunc",
	}
	funcObj := kubelessApi.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name:      funcName,
			Namespace: namespace,
			Labels:    funcLabels,
			UID:       "foo-uid",
		},
		Spec: kubelessApi.FunctionSpec{
			Function: "function",
			Deps:     "deps",
			Handler:  "foo.bar",
			Runtime:  "ruby2.4",
			Deployment: v1beta1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: funcAnno,
				},
				Spec: v1beta1.DeploymentSpec{
					Replicas: &replicas,
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
			},
		},
	}
	runtimeImages := `[
		{
			"ID": "ruby",
			"depName": "Gemfile",
			"fileNameSuffix": ".rb",
			"versions": [
				{
					"name": "ruby24",
					"version": "2.4",
					"initImage": "bitnami/ruby:2.4",
					"imagePullSecrets":[]
				}
			],
			"livenessProbeInfo":{
				"exec": {
					"command": [
						"curl",
						"-f",
						"http://localhost:8080/healthz"
					],
				},
				"initialDelaySeconds": 5,
				"periodSeconds": 10
			}
		}
	]`

	clientset := fake.NewSimpleClientset()
	kubelessConfigMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kubeless-config",
		},
		Data: map[string]string{"runtime-images": runtimeImages},
	}

	_, err := clientset.CoreV1().ConfigMaps(namespace).Create(kubelessConfigMap)
	if err != nil {
		logrus.Fatal("Unable to create configmap")
	}

	config, err := clientset.CoreV1().ConfigMaps(namespace).Get("kubeless-config", metav1.GetOptions{})
	if err != nil {
		logrus.Fatal("Unable to read the configmap")
	}
	var lr = langruntime.New(config)
	lr.ReadConfigMap()

	controller := FunctionController{
		logger:      logrus.WithField("pkg", "controller"),
		clientset:   clientset,
		langRuntime: lr,
		config:      config,
	}

	if err := controller.ensureK8sResources(&funcObj); err != nil {
		t.Fatalf("Creating/Updating resources returned err: %v", err)
	}
	dpm, _ := clientset.ExtensionsV1beta1().Deployments(namespace).Get(funcName, metav1.GetOptions{})
	expectedLivenessProbe := &v1.Probe{
		InitialDelaySeconds: int32(5),
		PeriodSeconds:       int32(10),
		Handler: v1.Handler{
			Exec: &v1.ExecAction{
				Command: []string{"curl", "-f", "http://localhost:8080/healthz"},
			},
		},
	}

	if !reflect.DeepEqual(dpm.Spec.Template.Spec.Containers[0].LivenessProbe, expectedLivenessProbe) {
		t.Fatalf("LivenessProbe found is '%v', although expected was '%v'", dpm.Spec.Template.Spec.Containers[0].LivenessProbe, expectedLivenessProbe)
	}

}
