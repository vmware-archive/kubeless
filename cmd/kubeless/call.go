/*
Copyright 2016 Skippbox, Ltd.

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
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/bitnami/kubeless/pkg/utils"
	"github.com/spf13/cobra"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/kubernetes/pkg/client/unversioned/remotecommand"
	k8scmd "k8s.io/kubernetes/pkg/kubectl/cmd"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

const (
	maxRetries       = 5
	defaultTimeSleep = 1 * time.Second
)

type defaultPortForwarder struct {
	cmdOut, cmdErr io.Writer
}

func (f *defaultPortForwarder) ForwardPorts(method string, url *url.URL, opts k8scmd.PortForwardOptions) error {
	dialer, err := remotecommand.NewExecutor(opts.Config, method, url)
	if err != nil {
		return err
	}
	fw, err := portforward.New(dialer, opts.Ports, opts.StopChannel, opts.ReadyChannel, f.cmdOut, f.cmdErr)
	if err != nil {
		return err
	}
	return fw.ForwardPorts()
}

var callCmd = &cobra.Command{
	Use:   "call <function_name> FLAG",
	Short: "call function from cli",
	Long:  `call function from cli`,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			jsonStr []byte
			get     bool = false
		)

		if len(args) != 1 {
			logrus.Fatal("Need exactly one argument - function name")
		}
		funcName := args[0]

		data, err := cmd.Flags().GetString("data")
		if data == "" {
			get = true
		} else {
			jsonStr = []byte(data)
		}

		if err != nil {
			logrus.Fatal(err)
		}
		ns, err := cmd.Flags().GetString("namespace")
		if err != nil {
			logrus.Fatal(err)
		}

		f := cmdutil.NewFactory(nil)
		k8sClient, err := f.RESTClient()
		if err != nil {
			logrus.Fatalf("Connection failed: %v", err)
		}
		clientset, err := f.ClientSet()
		if err != nil {
			logrus.Fatalf("Connection failed: %v", err)
		}
		k8sClientConfig, err := f.ClientConfig()
		if err != nil {
			logrus.Fatalf("Connection failed: %v", err)
		}

		//FIXME: we should only use restClient from client-go but now still have to use the old client for pf call
		k8sClientSet := utils.GetClientOutOfCluster()
		podName, err := utils.GetPodName(k8sClientSet, ns, funcName)
		if err != nil {
			logrus.Fatalf("Couldn't get the pod name: %v", err)
		}
		port, err := getLocalPort()
		if err != nil {
			logrus.Fatalf("Connection failed: %v", err)
		}
		portSlice := []string{port + ":8080"}

		go func() {
			pfo := k8scmd.PortForwardOptions{
				RESTClient: k8sClient,
				Namespace:  ns,
				Config:     k8sClientConfig,
				PodName:    podName,
				PodClient:  clientset.Core(),
				Ports:      portSlice,
				PortForwarder: &defaultPortForwarder{
					cmdOut: os.Stdout,
					cmdErr: os.Stderr,
				},
				StopChannel:  make(chan struct{}, 1),
				ReadyChannel: make(chan struct{}),
			}

			err = pfo.RunPortForward()
			if err != nil {
				logrus.Fatalf("Connection failed: %v", err)
			}

		}()

		retries := 0
		httpClient := &http.Client{}
		resp := &http.Response{}
		req := &http.Request{}
		url := fmt.Sprintf("http://localhost:%s", port)

		for {
			if get {
				req, _ = http.NewRequest("GET", url, nil)
			} else {
				req, _ = http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
				req.Header.Set("Content-Type", "application/json")
			}
			resp, err = httpClient.Do(req)

			if err == nil {
				htmlData, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					logrus.Fatalf("Response data is incorrect: %v", err)
				}
				fmt.Println(string(htmlData))
				resp.Body.Close()
				break
			} else {
				fmt.Println("Connecting to function...")
				retries++
				if retries == maxRetries {
					logrus.Fatalf("Can not connect to function. Please check your network connection!!!")
				}
				time.Sleep(defaultTimeSleep)
			}
		}
	},
}

func init() {
	callCmd.Flags().StringP("data", "", "", "Specify data for function")
	callCmd.Flags().StringP("namespace", "", api.NamespaceDefault, "Specify namespace for the function")

}

func getLocalPort() (string, error) {
	for i := 30000; i < 65535; i++ {
		conn, err := net.Dial("tcp", "127.0.0.1:"+strconv.Itoa(i))
		if err != nil {
			return strconv.Itoa(i), nil
		}
		conn.Close()
	}
	return "", errors.New("Can not find an unassigned port")
}
