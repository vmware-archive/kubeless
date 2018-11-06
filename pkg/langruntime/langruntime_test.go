package langruntime

import (
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"
)

var clientset = fake.NewSimpleClientset()

func TestMain(m *testing.M) {
	AddFakeConfig(clientset)
	os.Exit(m.Run())
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
	lr := SetupLangRuntime(clientset)
	lr.ReadConfigMap()

	expectedValues := []string{"requirements.txt", "test.py"}
	check(clientset, lr, "python2.7", "test", expectedValues, t)
}

func TestGetFunctionImage(t *testing.T) {
	lr := SetupLangRuntime(clientset)
	lr.ReadConfigMap()

	// Throws an error if the runtime doesn't exist
	_, err := lr.GetFunctionImage("unexistent")
	if err == nil {
		t.Fatalf("Retrieving data for 'unexistent' should return an error")
	}

	// Throws an error if the runtime version doesn't exist
	_, err = lr.GetFunctionImage("python10")
	expectedErrMsg := regexp.MustCompile("The given runtime and version python10 is not valid")
	if expectedErrMsg.FindString(err.Error()) == "" {
		t.Fatalf("Retrieving data for 'python10' should return an error. Received: %s", err)
	}

	expectedImageName := "ruby-test-image"
	os.Setenv("PYTHON2.7_RUNTIME", expectedImageName)
	imageR, errR := lr.GetFunctionImage("python2.7")
	if errR != nil {
		t.Errorf("Retrieving the image returned err: %v", errR)
	}
	if imageR != expectedImageName {
		t.Errorf("Expecting " + imageR + " to be set to " + expectedImageName)
	}
	os.Unsetenv("PYTHON2.7_RUNTIME")
}

func TestGetLivenessProbe(t *testing.T) {
	lr := SetupLangRuntime(clientset)
	lr.ReadConfigMap()
	livenessProbe := lr.GetLivenessProbeInfo("python", 8080)

	expectedLivenessProbe := &v1.Probe{
		InitialDelaySeconds: int32(5),
		PeriodSeconds:       int32(10),
		Handler: v1.Handler{
			Exec: &v1.ExecAction{
				Command: []string{"curl", "-f", "http://localhost:8080/healthz"},
			},
		},
	}

	if !reflect.DeepEqual(livenessProbe, expectedLivenessProbe) {
		t.Fatalf("Expected livenessProbeInfo to be %v, but found %v", expectedLivenessProbe, livenessProbe)
	}
}

func TestGetRuntimes(t *testing.T) {
	lr := SetupLangRuntime(clientset)
	lr.ReadConfigMap()

	runtimes := strings.Join(lr.GetRuntimes(), ", ")
	expectedRuntimes := "python2.7"
	if runtimes != expectedRuntimes {
		t.Errorf("Expected %s but got %s", expectedRuntimes, runtimes)
	}
}

func TestGetBuildContainer(t *testing.T) {
	lr := SetupLangRuntime(clientset)
	lr.ReadConfigMap()

	// It should throw an error if there is not an image available
	_, err := lr.GetBuildContainer("notExists", "", []v1.EnvVar{}, v1.VolumeMount{})
	if err == nil {
		t.Error("Expected to throw an error")
	}

	// It should return the proper build image for python
	vol1 := v1.VolumeMount{Name: "v1", MountPath: "/v1"}
	c, err := lr.GetBuildContainer("python2.7", "abc123", []v1.EnvVar{}, vol1)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	expectedContainer := v1.Container{
		Name:            "install",
		Image:           "python:2.7",
		Command:         []string{"sh", "-c"},
		Args:            []string{"echo 'abc123  /v1/requirements.txt' > /tmp/deps.sha256 && sha256sum -c /tmp/deps.sha256 && foo"},
		VolumeMounts:    []v1.VolumeMount{vol1},
		WorkingDir:      "/v1",
		ImagePullPolicy: v1.PullIfNotPresent,
		Env: []v1.EnvVar{
			{Name: "KUBELESS_INSTALL_VOLUME", Value: "/v1"},
			{Name: "KUBELESS_DEPS_FILE", Value: "/v1/requirements.txt"},
		},
	}
	if !reflect.DeepEqual(expectedContainer, c) {
		t.Errorf("Unexpected result. Expecting:\n %+v\nReceived:\n %+v", expectedContainer, c)
	}
}
