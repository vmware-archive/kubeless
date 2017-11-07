/*
Copyright (c) 2016-2017 Bitnami

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path"
	"time"

	"github.com/kubeless/kubeless/pkg/utils"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
)

func postFile(filename string, url string, restCli rest.RESTClient) error {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	// Create writer from "FormFile"
	fileWriter, err := bodyWriter.CreateFormFile("uploadfile", filename)
	if err != nil {
		fmt.Println("error writing to buffer")
		return err
	}

	// Open file to upload
	fh, err := os.Open(filename)
	if err != nil {
		fmt.Println("error opening file")
		return err
	}
	defer fh.Close()

	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		return err
	}

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	// POST content
	_, err = restCli.Post().AbsPath(url).Body(bodyBuf).SetHeader("Content-Type", contentType).Do().Raw()
	if err != nil {
		return err
	}
	return nil
}

// UploadFunction uploads a file to Minio using as object name file.extension
// It uses a Kubernetes job to access Minio since we need to use an URL <domain>[:port] (a URL with
// proxy is not valid)
func uploadFunction(file, checksum, namespace string, fileStats os.FileInfo, timeout int, cli kubernetes.Interface, restCli rest.RESTClient) (string, error) {
	uploadName := "upload-file-" + checksum[0:10]
	fileName := path.Base(file) + "." + checksum
	var absPath string
	if !path.IsAbs(file) {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		absPath = path.Join(cwd, file)
	} else {
		absPath = file
	}
	// Define the SVC to expose
	svc := v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      uploadName,
			Namespace: namespace,
			Labels: map[string]string{
				"kubeless": "proxy-uploader",
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Port:       8080,
					TargetPort: intstr.FromInt(8080),
					NodePort:   0,
					Protocol:   v1.ProtocolTCP,
				},
			},
			Selector: map[string]string{
				"kubeless": "proxy-uploader",
			},
			Type: v1.ServiceTypeClusterIP,
		},
	}
	minioCredentials := "minio-key"
	// Define the POD to run
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      uploadName,
			Namespace: namespace,
			Labels: map[string]string{
				"kubeless": "proxy-uploader",
			},
		},
		Spec: v1.PodSpec{
			Volumes: []v1.Volume{
				{
					Name: minioCredentials,
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							SecretName: minioCredentials,
						},
					},
				},
			},
			RestartPolicy: v1.RestartPolicyNever,
			Containers: []v1.Container{
				{
					Name:  "uploader",
					Image: "kubeless/proxy-uploader:ffu",
					Env: []v1.EnvVar{
						v1.EnvVar{
							Name:  "MINIO_NAMESPACE",
							Value: namespace,
						},
					},
					VolumeMounts: []v1.VolumeMount{
						{
							Name:      minioCredentials,
							MountPath: "/minio-cred",
						},
					},
					Command: []string{"proxy-uploader", "-d", "/tmp/"},
				},
			},
		},
	}
	_, err := cli.CoreV1().Pods(namespace).Create(&pod)
	if err != nil {
		return "", err
	}
	defer cli.CoreV1().Pods(namespace).Delete(uploadName, &metav1.DeleteOptions{})
	createdSvc, err := cli.CoreV1().Services(namespace).Create(&svc)
	if err != nil {
		return "", err
	}
	defer cli.CoreV1().Services(namespace).Delete(uploadName, &metav1.DeleteOptions{})

	// Wait for the endpoint to be ready
	wait.Poll(time.Duration(time.Second), time.Duration(10)*time.Second, func() (bool, error) {
		fmt.Printf("Waiting for the endpoint %s to be ready\n", uploadName)
		ep, err := cli.CoreV1().Endpoints(namespace).Get(uploadName, metav1.GetOptions{})
		if err != nil {
			if k8sErrors.IsNotFound(err) {
				return false, nil
			}
			return false, err
		}
		if len(ep.Subsets) > 0 && ep.Subsets[0].Addresses[0].IP != "" {
			fmt.Println(ep.Subsets)
			return true, nil
		}
		return false, nil
	})

	// Upload the file
	url := createdSvc.ObjectMeta.SelfLink + "/proxy/upload"
	err = postFile(absPath, url, restCli)
	if err != nil {
		return "", err
	}
	return "http://minio." + namespace + ":9000/functions/" + fileName, nil
}

func main() {
	restCli, err := utils.GetRestClientOutOfCluster("", "v1", "/api")
	if err != nil {
		panic(err)
	}
	stats, _ := os.Stat("/Users/andresmartinez/projects/src/github.com/kubeless/kubeless/examples/nodejs/helloFunctions.zip")
	cli := utils.GetClientOutOfCluster()
	url, err := uploadFunction(
		"/Users/andresmartinez/projects/src/github.com/kubeless/kubeless/examples/nodejs/helloFunctions.zip",
		"sfjsrgliemgt3w490ptfjnmelktn098e3jr4234tmf09pg5ope3gg",
		"kubeless",
		stats,
		120,
		cli,
		*restCli,
	)
	if err != nil {
		panic(err)
	}
	fmt.Println("URL! " + url)
}
