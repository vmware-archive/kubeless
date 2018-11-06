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
      "compiled": false,
      "depName": "requirements.txt",
			"fileNameSuffix": ".py",
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
            "images": [
               {
                  "command": "foo",
                  "image": "python:2.7",
                  "phase": "installation"
               },
               {
                  "image": "bar",
									"phase": "runtime",
									"env": {
										"PYTHONPATH": "/kubeless/lib/python2.7/site-packages:/kubeless"
									}
               }
            ],
            "name": "python27",
						"version": "2.7",
						"imagePullSecrets": [{"ImageSecret": "p1"}, {"ImageSecret": "p2"}]
        },
	  	]
		}
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
