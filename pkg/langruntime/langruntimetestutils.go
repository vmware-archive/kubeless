package langruntime

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// AddFakeConfig initializes configmap for unit tests with fake configuration.
func AddFakeConfig(clientset *fake.Clientset) {

	runtimeImages := `[
		{
			"ID": "python",
			"depName": "requirements.txt",
			"versions": [
				{
					"name": "python27",
					"version": "2.7",
					"initImage": "tuna/python-pillow:2.7.11-alpine",
					"runtimeImage": "kubeless/python@sha256:0f3b64b654df5326198e481cd26e73ecccd905aae60810fc9baea4dcbb61f697",
					"imagePullSecrets": [{"ImageSecret": "p1"}, {"ImageSecret": "p2"}]
				},
			],
			"fileNameSuffix": ".py"
		},
		{
			"ID": "nodejs",
			"depName": "package.json",
			"livenessProbeInfo": {
				"exec": {
					"command": [
						"curl",
						"-f",
						"http://localhost:8080/healthz"
					]
				},
				"initialDelaySeconds": 5,
				"periodseconds": 10
			},
			"versions": [
				{
					"name": "nodejs6",
					"version": "6",
					"initImage": "node:6.10",
					"imagePullSecrets": [{"ImageSecret": "p1"}, {"ImageSecret": "p2"}]
				}
			],
			"fileNameSuffix": ".js"
		},
		{
			"ID": "ruby",
			"depName": "Gemfile",
			"versions": [
				{
					"name": "ruby24",
					"version": "2.4",
					InitImage: "bitnami/ruby:2.4",
					"imagePullSecrets": [{"ImageSecret": "p1"}, {"ImageSecret": "p2"}]
				},
			],
			"fileNameSuffix": ".rb"
		},
	]`

	cm := v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kubeless-config",
			Namespace: "kubeless",
		},
		Data: map[string]string{
			"runtime-images": runtimeImages,
		},
	}

	_, err := clientset.CoreV1().ConfigMaps("kubeless").Create(&cm)
	if err != nil {
		logrus.Fatal("Unable to create configmap")
	}
}

// SetupLangRuntime Sets up Langruntime struct
func SetupLangRuntime(clientset *fake.Clientset) *Langruntimes {
	config, err := clientset.CoreV1().ConfigMaps("kubeless").Get("kubeless-config", metav1.GetOptions{})
	if err != nil {
		logrus.Fatal("Unable to read the configmap")
	}
	var lr = New(config)
	return lr
}
