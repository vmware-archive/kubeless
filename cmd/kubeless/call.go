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
	"github.com/skippbox/kubeless/pkg/utils"
	"github.com/spf13/cobra"
	"k8s.io/kubernetes/pkg/client/unversioned/portforward"
	"k8s.io/kubernetes/pkg/client/unversioned/remotecommand"
	k8scmd "k8s.io/kubernetes/pkg/kubectl/cmd"
)

const (
	MaxRetries       = 5
	DefaultTimeSleep = 1 * time.Second
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
		if err != nil {
			logrus.Fatal(err)
		}
		ns, err := cmd.Flags().GetString("namespace")
		if data == "" {
			get = true
		} else {
			jsonStr = []byte(data)
		}

		f := utils.GetFactory()
		if ns == "" {
			ns, _, err = f.DefaultNamespace()
			if err != nil {
				logrus.Fatalf("Connection failed: %v", err)
			}
		}
		kClient, err := f.Client()
		if err != nil {
			logrus.Fatalf("Connection failed: %v", err)
		}
		kClientConfig, err := f.ClientConfig()
		if err != nil {
			logrus.Fatalf("Connection failed: %v", err)
		}

		podName, err := utils.GetPodName(kClient, ns, funcName)
		port, err := getLocalPort()
		if err != nil {
			logrus.Fatalf("Connection failed: %v", err)
		}
		portSlice := []string{port + ":8080"}

		go func() {
			pfo := k8scmd.PortForwardOptions{
				Client:    kClient,
				Namespace: ns,
				Config:    kClientConfig,
				PodName:   podName,
				Ports:     portSlice,
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
				retries += 1
				if retries == MaxRetries {
					logrus.Fatalf("Can not connect to function. Please check your network connection!!!")
				}
				time.Sleep(DefaultTimeSleep)
			}
		}
	},
}

func init() {
	callCmd.Flags().StringP("data", "", "", "Specify data for function")
	callCmd.Flags().StringP("namespace", "", "", "Specify namespace for the function")

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
