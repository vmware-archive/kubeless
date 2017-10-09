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
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/kubeless/kubeless/pkg/utils"
	"github.com/spf13/cobra"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/kubernetes/pkg/client/unversioned/remotecommand"
	k8scmd "k8s.io/kubernetes/pkg/kubectl/cmd"
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

		tprClient, err := utils.GetTPRClientOutOfCluster()
		url := "/api/v1/proxy/namespaces/" + ns + "/services/" + funcName + "/"

		req := &rest.Request{}
		if get {
			req = tprClient.Get().AbsPath(url)
		} else {
			req = tprClient.Post().AbsPath(url).Body(bytes.NewBuffer(jsonStr)).SetHeader("Content-Type", "application/json")
		}
		res, err := req.Do().Raw()
		if err != nil {
			// Properly interpret line breaks
			logrus.Fatal(strings.Replace(err.Error(), `\n`, "\n", -1))
		}
		fmt.Println(string(res))
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
