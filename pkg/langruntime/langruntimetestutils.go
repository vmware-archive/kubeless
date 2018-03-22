package langruntime

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// AddFakeConfig initializes configmap for unit tests with fake configuration.
func AddFakeConfig(clientset *fake.Clientset) {

	var runtimeImages = []RuntimeInfo{{
		ID:             "python",
		DepName:        "requirements.txt",
		FileNameSuffix: ".py",
		Versions: []RuntimeVersion{
			{
				Name:         "python27",
				Version:      "2.7",
				InitImage:    "tuna/python-pillow:2.7.11-alpine",
				RuntimeImage: "kubeless/python@sha256:0f3b64b654df5326198e481cd26e73ecccd905aae60810fc9baea4dcbb61f697",

				ImagePullSecrets: []ImageSecret{
					{ImageSecret: "p1"}, {ImageSecret: "p2"},
				},
			}, {
				Name:    "python34",
				Version: "3.4",
				ImagePullSecrets: []ImageSecret{
					{ImageSecret: "p1"}, {ImageSecret: "p2"},
				},
			}, {
				Name:    "python36",
				Version: "3.6",
				ImagePullSecrets: []ImageSecret{
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
				ImagePullSecrets: []ImageSecret{
					{ImageSecret: "p1"}, {ImageSecret: "p2"},
				},
			}, {
				Name:    "nodejs8",
				Version: "8",
				ImagePullSecrets: []ImageSecret{
					{ImageSecret: "p1"}, {ImageSecret: "p2"},
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
				ImagePullSecrets: []ImageSecret{
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
				ImagePullSecrets: []ImageSecret{
					{ImageSecret: "p1"}, {ImageSecret: "p2"},
				},
			},
		},
	}, {ID: "php",
		DepName:        "composer.json",
		FileNameSuffix: ".php",
		Versions: []RuntimeVersion{
			{
				Name:      "php7.2",
				Version:   "7.2",
				InitImage: "composer:1.6",
				ImagePullSecrets: []ImageSecret{
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
