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
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/kubeless/kubeless/pkg/spec"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	batchv1 "k8s.io/client-go/pkg/apis/batch/v1"
)

var functionCmd = &cobra.Command{
	Use:   "function SUBCOMMAND",
	Short: "function specific operations",
	Long:  `function command allows user to list, deploy, edit, delete functions running on Kubeless`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	functionCmd.AddCommand(deployCmd)
	functionCmd.AddCommand(deleteCmd)
	functionCmd.AddCommand(listCmd)
	functionCmd.AddCommand(callCmd)
	functionCmd.AddCommand(logsCmd)
	functionCmd.AddCommand(describeCmd)
	functionCmd.AddCommand(updateCmd)
	functionCmd.AddCommand(autoscaleCmd)
}

func getKV(input string) (string, string) {
	var key, value string
	if pos := strings.IndexAny(input, "=:"); pos != -1 {
		key = input[:pos]
		value = input[pos+1:]
	} else {
		// no separator found
		key = input
		value = ""
	}

	return key, value
}

func parseLabel(labels []string) map[string]string {
	funcLabels := map[string]string{}
	for _, label := range labels {
		k, v := getKV(label)
		funcLabels[k] = v
	}
	return funcLabels
}

func parseEnv(envs []string) []v1.EnvVar {
	funcEnv := []v1.EnvVar{}
	for _, env := range envs {
		k, v := getKV(env)
		funcEnv = append(funcEnv, v1.EnvVar{
			Name:  k,
			Value: v,
		})
	}
	return funcEnv
}

func parseMemory(mem string) (resource.Quantity, error) {
	quantity, err := resource.ParseQuantity(mem)
	if err != nil {
		return resource.Quantity{}, err
	}

	return quantity, nil
}

func getFileSha256(file string) (string, error) {
	var checksum string
	h := sha256.New()
	ff, err := os.Open(file)
	if err != nil {
		return checksum, err
	}
	defer ff.Close()
	_, err = io.Copy(h, ff)
	if err != nil {
		return checksum, err
	}
	checksum = hex.EncodeToString(h.Sum(nil))
	return checksum, err
}

func waitForCompletedJob(jobName, namespace string, timeout int, cli kubernetes.Interface) error {
	counter := 0
	for counter < timeout {
		j, err := cli.BatchV1().Jobs(namespace).Get(jobName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if j.Status.Failed == 1 {
			err = fmt.Errorf("Unable to run upload job. Received: %s", j.Status.Conditions[0].Message)
			return err
		} else if j.Status.Succeeded == 1 {
			return nil
		}
		time.Sleep(time.Duration(time.Second))
		counter++
	}
	return fmt.Errorf("Upload job has not finished after %s seconds", string(timeout))
}

func uploadFunctionToMinio(file, checksum string, cli kubernetes.Interface) (string, error) {
	minioCredentials := "minio-key"
	jobName := "upload-file"
	fileName := path.Base(file) + "." + checksum
	job := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: "kubeless",
		},
		Spec: batchv1.JobSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						v1.Volume{
							Name: minioCredentials,
							VolumeSource: v1.VolumeSource{
								Secret: &v1.SecretVolumeSource{
									SecretName: minioCredentials,
								},
							},
						},
						v1.Volume{
							Name: "func",
							VolumeSource: v1.VolumeSource{
								HostPath: &v1.HostPathVolumeSource{
									Path: file,
								},
							},
						},
					},
					RestartPolicy: v1.RestartPolicyNever,
					Containers: []v1.Container{
						v1.Container{
							Name:  "uploader",
							Image: "minio/mc:RELEASE.2017-10-14T00-51-16Z",
							VolumeMounts: []v1.VolumeMount{
								v1.VolumeMount{
									Name:      minioCredentials,
									MountPath: "/minio-cred",
								},
								v1.VolumeMount{
									Name:      "func",
									MountPath: "/" + path.Base(file),
								},
							},
							Command: []string{"sh", "-c"},
							Args: []string{
								"mc config host add minioserver http://minio.kubeless:9000 $(cat /minio-cred/accesskey) $(cat /minio-cred/secretkey); " +
									"mc cp /" + path.Base(file) + " minioserver/functions/" + fileName,
							},
						},
					},
				},
			},
		},
	}
	_, err := cli.BatchV1().Jobs("kubeless").Create(&job)
	if err != nil {
		return "", err
	}
	err = waitForCompletedJob(jobName, "kubeless", 120, cli)
	if err != nil {
		return "", err
	}
	// Clean up (delete job)
	err = cli.BatchV1().Jobs("kubeless").Delete(jobName, &metav1.DeleteOptions{})
	if err != nil {
		return "", err
	}
	return "http://minio.kubeless:9000/functions/" + fileName, nil
}

// Convert MB to Bytes as int64
func mbtoInt64(mb int) int64 {
	return int64(mb * 1024 * 1024)
}

func uploadFunction(file string, cli kubernetes.Interface) (string, string, error) {
	var checksum string
	stats, err := os.Stat(file)
	if stats.Size() > mbtoInt64(50) { // TODO: Make the max file size configurable
		err = errors.New("The maximum size of a function is 50MB")
		return "", "", err
	}
	checksum, err = getFileSha256(file)
	if err != nil {
		return "", "", err
	}
	url, err := uploadFunctionToMinio(file, checksum, cli)
	if err != nil {
		return "", "", err
	}
	return url, checksum, err
}

func isMinioAvailable(cli kubernetes.Interface) bool {
	_, err := cli.Core().Services("kubeless").Get("minio", metav1.GetOptions{})
	if err != nil {
		return false
	}
	return true
}

func getFunctionDescription(funcName, ns, handler, file, deps, runtime, topic, schedule, runtimeImage, mem string, triggerHTTP bool, envs, labels []string, defaultFunction spec.Function, cli kubernetes.Interface) (*spec.Function, error) {

	if handler == "" {
		handler = defaultFunction.Spec.Handler
	}

	var function, checksum, contentType string
	if file == "" {
		file = defaultFunction.Spec.File
		contentType = defaultFunction.Spec.ContentType
		function = defaultFunction.Spec.Function
		checksum = defaultFunction.Spec.Checksum
	} else {
		var err error
		if isMinioAvailable(cli) {
			function, checksum, err = uploadFunction(file, cli)
			contentType = "URL"
			if err != nil {
				return &spec.Function{}, err
			}
		} else {
			// If an object storage service is not available check
			// that the file is not over 1MB to store it as a Custom Resource
			stats, err := os.Stat(file)
			if err != nil {
				return &spec.Function{}, err
			}
			if stats.Size() > mbtoInt64(1) {
				err = errors.New("Unable to deploy functions over 1MB withouth a storage service")
				return &spec.Function{}, err
			}
			functionBytes, err := ioutil.ReadFile(file)
			if err != nil {
				return &spec.Function{}, err
			}
			if err != nil {
				return &spec.Function{}, err
			}
			fileType := http.DetectContentType(functionBytes)
			if strings.Contains(fileType, "text/plain") {
				function = string(functionBytes[:])
				contentType = "text"
			} else {
				function = base64.StdEncoding.EncodeToString(functionBytes)
				contentType = "base64"
			}
			c, err := getFileSha256(file)
			checksum = "sha256:" + c
			if err != nil {
				return &spec.Function{}, err
			}
		}
	}

	if deps == "" {
		deps = defaultFunction.Spec.Deps
	}

	if runtime == "" {
		runtime = defaultFunction.Spec.Runtime
	}

	funcType := ""
	switch {
	case triggerHTTP:
		funcType = "HTTP"
		topic = ""
		schedule = ""
		break
	case schedule != "":
		funcType = "Scheduled"
		topic = ""
		break
	case topic != "":
		funcType = "PubSub"
		schedule = ""
		break
	default:
		funcType = defaultFunction.Spec.Type
		topic = defaultFunction.Spec.Topic
		schedule = defaultFunction.Spec.Schedule
	}

	funcEnv := parseEnv(envs)
	if len(funcEnv) == 0 && len(defaultFunction.Spec.Template.Spec.Containers) != 0 {
		funcEnv = defaultFunction.Spec.Template.Spec.Containers[0].Env
	}

	funcLabels := parseLabel(labels)
	if len(funcLabels) == 0 {
		funcLabels = defaultFunction.Metadata.Labels
	}

	resources := v1.ResourceRequirements{}
	if mem != "" {
		funcMem, err := parseMemory(mem)
		if err != nil {
			err := fmt.Errorf("Wrong format of the memory value: %v", err)
			return &spec.Function{}, err
		}
		resource := map[v1.ResourceName]resource.Quantity{
			v1.ResourceMemory: funcMem,
		}
		resources = v1.ResourceRequirements{
			Limits:   resource,
			Requests: resource,
		}
	} else {
		if len(defaultFunction.Spec.Template.Spec.Containers) != 0 {
			resources = defaultFunction.Spec.Template.Spec.Containers[0].Resources
		}
	}

	if len(runtimeImage) == 0 && len(defaultFunction.Spec.Template.Spec.Containers) != 0 {
		runtimeImage = defaultFunction.Spec.Template.Spec.Containers[0].Image
	}

	return &spec.Function{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Function",
			APIVersion: "k8s.io/v1",
		},
		Metadata: metav1.ObjectMeta{
			Name:      funcName,
			Namespace: ns,
			Labels:    funcLabels,
		},
		Spec: spec.FunctionSpec{
			Handler:     handler,
			Runtime:     runtime,
			Type:        funcType,
			Function:    function,
			File:        path.Base(file),
			Checksum:    checksum,
			ContentType: contentType,
			Deps:        deps,
			Topic:       topic,
			Schedule:    schedule,
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Env:       funcEnv,
							Resources: resources,
							Image:     runtimeImage,
						},
					},
				},
			},
		},
	}, nil
}
