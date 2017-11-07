package main

import (
	"os"

	"github.com/kubeless/kubeless/pkg/utils"
	"github.com/minio/minio-go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func cleanFunctionsByNumber(minioCli minio.Client, max int) error {

}

func main() {
	clientset := utils.GetClient()
	minioSVC, err := clientset.CoreV1().Services("kubeless").Get("minio", metav1.GetOptions{})
	if err != nil {
		panic(err)
	}
	endpoint := "minio.kubeless:" + minioSVC.Spec.Ports[0].Port
	minioSecret, err := clientset.CoreV1().Secrets("kubeless").Get("minio-key", metav1.GetOptions{})
	if err != nil {
		panic(err)
	}
	accessKeyID := string(minioSecret.Data["accesskey"])
	secretAccessKey := string(minioSecret.Data["secretkey"])
	useSSL := false

	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, accessKeyID, secretAccessKey, useSSL)
	if err != nil {
		panic(err)
	}

	minioConfig, err := clientset.CoreV1().ConfigMaps("kubeless").Get("minio-config")
	if err != nil {
		panic(err)
	}
	cleanPolicy := os.Getenv("CLEAN_POLICY")
	if cleanPolicy == "" {
		cleanPolicy = "maxFuncHistory"
	}

	switch cleanPolicy {
	case "maxFuncHistory":

	}
	// log.Printf("%#v\n", minioClient) // minioClient is now setup
}
