package langruntime

import (
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"

	yaml "github.com/ghodss/yaml"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func initializeConfigmap(clientset *fake.Clientset, lr *Langruntimes) {

	var runtimeImages = []RuntimeInfo{{
		ID:             "python",
		DepName:        "requirements.txt",
		FileNameSuffix: ".py",
		Versions: []RuntimeVersion{
			{
				Name:      "python27",
				Version:   "2.7",
				InitImage: "tuna/python-pillow:2.7.11-alpine",
				ImagePullSecrets: []ImageSecrets{
					{ImageSecret: "p1"}, {ImageSecret: "p2"},
				},
			}, {
				Name:    "python34",
				Version: "3.4",
				ImagePullSecrets: []ImageSecrets{
					{ImageSecret: "p1"}, {ImageSecret: "p2"},
				},
			}, {
				Name:    "python36",
				Version: "3.6",
				ImagePullSecrets: []ImageSecrets{
					{ImageSecret: "p1"}, {ImageSecret: "p2"},
				},
			},
		},
	}, {ID: "nodejs",
		DepName:        "package.json",
		FileNameSuffix: ".js",
		Versions: []RuntimeVersion{
			{
				Name:      "nodejs6",
				Version:   "6",
				InitImage: "node:6.10",
				ImagePullSecrets: []ImageSecrets{
					{ImageSecret: "p1"}, {ImageSecret: "p2"},
				},
			}, {
				Name:    "nodejs8",
				Version: "8",
				ImagePullSecrets: []ImageSecrets{
					{ImageSecret: "p1"}, {"p2"},
				},
			},
		},
	}, {ID: "ruby",
		DepName:        "Gemfile",
		FileNameSuffix: ".rb",
		Versions: []RuntimeVersion{
			{
				Name:      "ruby24",
				Version:   "2.4",
				InitImage: "bitnami/ruby:2.4",
				ImagePullSecrets: []ImageSecrets{
					{ImageSecret: "p1"}, {ImageSecret: "p2"},
				},
			},
		},
	}, {ID: "dotnetcore",
		DepName:        "requirements.xml",
		FileNameSuffix: ".cs",
		Versions: []RuntimeVersion{
			{
				Name:    "dotnetcore2.0",
				Version: "2.0",
				ImagePullSecrets: []ImageSecrets{
					{ImageSecret: "p1"}, {ImageSecret: "p2"},
				},
			},
		},
	}}

	out, err := yaml.Marshal(runtimeImages)
	if err != nil {
		logrus.Fatal("Canot Marshall runtimeimage")
	}

	cm := v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kubeless-config",
			Namespace: "kubeless",
		},
		Data: map[string]string{
			"runtime-images": string(out),
		},
	}

	_, err = clientset.CoreV1().ConfigMaps("kubeless").Create(&cm)
	if err != nil {
		logrus.Fatal("Unable to create configmap")
	}

	lr.ReadConfigMap()

}

func check(clientset *fake.Clientset, lr *Langruntimes, runtime, fname string, values []string, t *testing.T) {

	info, err := lr.GetRuntimeInfo(runtime)
	if err != nil {
		t.Fatal(err)
	}
	if info.DepName != values[0] {
		t.Fatalf("Retrieving the image returned a wrong dependencies file. Received " + info.DepName + " while expecting " + values[0])
	}
	if fname+info.FileNameSuffix != values[1] {
		t.Fatalf("Retrieving the image returned a wrong file name. Received " + fname + info.FileNameSuffix + " while expecting " + values[1])
	}
}

func TestGetFunctionFileNames(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	var lr = New(clientset, "kubeless", "kubeless-config")
	initializeConfigmap(clientset, lr)

	expectedValues := []string{"requirements.txt", "test.py"}
	check(clientset, lr, "python2.7", "test", expectedValues, t)
	check(clientset, lr, "python3.4", "test", expectedValues, t)

	expectedValues = []string{"package.json", "test.js"}
	check(clientset, lr, "nodejs6", "test", expectedValues, t)
	check(clientset, lr, "nodejs8", "test", expectedValues, t)

	expectedValues = []string{"Gemfile", "test.rb"}
	check(clientset, lr, "ruby2.4", "test", expectedValues, t)

	expectedValues = []string{"requirements.xml", "test.cs"}
	check(clientset, lr, "dotnetcore2.0", "test", expectedValues, t)
}

func TestGetFunctionImage(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	var lr = New(clientset, "kubeless", "kubeless-config")
	initializeConfigmap(clientset, lr)

	// Throws an error if the runtime doesn't exist
	_, err := lr.GetFunctionImage("unexistent", "HTTP")
	if err == nil {
		t.Fatalf("Retrieving data for 'unexistent' should return an error")
	}

	// Throws an error if the runtime version doesn't exist
	_, err = lr.GetFunctionImage("nodejs3", "HTTP")
	expectedErrMsg := regexp.MustCompile("The given runtime and version nodejs3 is not valid")
	if expectedErrMsg.FindString(err.Error()) == "" {
		t.Fatalf("Retrieving data for 'nodejs3' should return an error. Received: %s", err)
	}

	expectedImageName := "ruby-test-image"
	os.Setenv("RUBY2.4_RUNTIME", expectedImageName)
	imageR, errR := lr.GetFunctionImage("ruby2.4", "HTTP")
	if errR != nil {
		t.Fatalf("Retrieving the image returned err: %v", errR)
	}
	if imageR != expectedImageName {
		t.Fatalf("Expecting " + imageR + " to be set to " + expectedImageName)
	}
	os.Unsetenv("RUBY2.4_RUNTIME")

	expectedImageName = "ruby-pubsub-test-image"
	os.Setenv("RUBY2.4_PUBSUB_RUNTIME", "ruby-pubsub-test-image")
	imageR, errR = lr.GetFunctionImage("ruby2.4", "PubSub")
	if errR != nil {
		t.Fatalf("Retrieving the image returned err: %v", errR)
	}
	if imageR != expectedImageName {
		t.Fatalf("Expecting " + imageR + " to be set to " + expectedImageName)
	}
	os.Unsetenv("RUBY2.4_PUBSUB_RUNTIME")
}

func TestGetRuntimes(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	var lr = New(clientset, "kubeless", "kubeless-config")
	initializeConfigmap(clientset, lr)

	runtimes := strings.Join(lr.GetRuntimes(), ", ")
	expectedRuntimes := "python2.7, python3.4, python3.6, nodejs6, nodejs8, ruby2.4, dotnetcore2.0"
	if runtimes != expectedRuntimes {
		t.Errorf("Expected %s but got %s", expectedRuntimes, runtimes)
	}
}

func TestGetBuildContainer(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	var lr = New(clientset, "kubeless", "kubeless-config")
	initializeConfigmap(clientset, lr)
	// It should throw an error if there is not an image available
	_, err := lr.GetBuildContainer("notExists", []v1.EnvVar{}, v1.VolumeMount{})
	if err == nil {
		t.Error("Expected to throw an error")
	}

	// It should return the proper build image for python
	env := []v1.EnvVar{}
	vol1 := v1.VolumeMount{Name: "v1", MountPath: "/v1"}
	c, err := lr.GetBuildContainer("python2.7", env, vol1)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	expectedContainer := v1.Container{
		Name:            "install",
		Image:           "tuna/python-pillow:2.7.11-alpine",
		Command:         []string{"sh", "-c"},
		Args:            []string{"pip install --prefix=/v1 -r /v1/requirements.txt"},
		VolumeMounts:    []v1.VolumeMount{vol1},
		WorkingDir:      "/v1",
		ImagePullPolicy: v1.PullIfNotPresent,
		Env:             env,
	}
	if !reflect.DeepEqual(expectedContainer, c) {
		t.Errorf("Unexpected result. Expecting:\n %+v\nReceived:\n %+v", expectedContainer, c)
	}

	// It should return the proper build image for nodejs
	nodeEnv := []v1.EnvVar{
		{Name: "NPM_REGISTRY", Value: "http://reg.com"},
		{Name: "NPM_SCOPE", Value: "myorg"},
	}
	c, err = lr.GetBuildContainer("nodejs6", nodeEnv, vol1)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if c.Image != "node:6.10" {
		t.Errorf("Unexpected image %s", c.Image)
	}
	if !strings.Contains(c.Args[0], "npm config set myorg:registry http://reg.com && npm install") {
		t.Errorf("Unexpected command %s", c.Args[0])
	}

	// It should return the proper build image for ruby
	c, err = lr.GetBuildContainer("ruby2.4", env, vol1)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if c.Image != "bitnami/ruby:2.4" {
		t.Errorf("Unexpected image %s", c.Image)
	}
	if !strings.Contains(c.Args[0], "bundle install --gemfile=/v1/Gemfile --path=/v1") {
		t.Errorf("Unexpected command %s", c.Args[0])
	}

}
