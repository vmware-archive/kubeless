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
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/minio/minio-go"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/pkg/api/v1"
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

func readFile(file string) (string, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}
	return string(data[:]), nil
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

func getFileSha256(file string) (sha256Sum [32]byte, checksum string, err error) {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}
	h := sha256.New()
	ff, err := os.Open(file)
	if err != nil {
		return
	}
	defer ff.Close()
	_, err = io.Copy(h, ff)
	if err != nil {
		return
	}
	sha256Sum = sha256.Sum256(content)
	checksum = hex.EncodeToString(h.Sum(nil))
	return
}

func uploadFunction(file string) (checksum string, err error) {

	stats, err := os.Stat(file)
	if stats.Size() > int64(52428800) { // TODO: Make the max file size configurable
		err = errors.New("The maximum size of a function is 50MB")
		return
	}
	sha256Sum, checksum, err := getFileSha256(file)

	endpoint := "192.168.99.100:30276"
	accessKeyID := "foobar"
	secretAccessKey := "foobarfoo"
	useSSL := false
	minioClient, err := minio.New(endpoint, accessKeyID, secretAccessKey, useSSL)
	if err != nil {
		return
	}
	bucketName := "functions"
	objectName := path.Base(file) + "." + checksum

	_, getObjectErr := minioClient.StatObject(bucketName, objectName, minio.StatObjectOptions{})
	if getObjectErr == nil {
		// File already exists, validate checksum
		var dir string
		dir, err = ioutil.TempDir("", "")
		if err != nil {
			return
		}
		defer os.RemoveAll(dir)

		err = minioClient.FGetObject(bucketName, objectName, path.Join(dir, objectName), minio.GetObjectOptions{})
		if err != nil {
			return
		}
		var uploadedContent []byte
		uploadedContent, err = ioutil.ReadFile(path.Join(dir, objectName))
		if err != nil {
			return
		}
		uploadedSha256 := sha256.Sum256(uploadedContent)
		if sha256Sum != uploadedSha256 {
			err = fmt.Errorf("The function %s has an invalid checksum %s", objectName, uploadedSha256)
			return
		}
		logrus.Info("Skipping function storage since %s is already present", file)
	} else {
		// Upload the yaml file with FPutObject
		_, err = minioClient.FPutObject(bucketName, objectName, file, minio.PutObjectOptions{})
		if err != nil {
			return
		}
	}
	return
}
