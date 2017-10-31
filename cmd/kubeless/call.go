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
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/kubeless/kubeless/pkg/utils"
	"github.com/spf13/cobra"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
)

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

		crdClient, err := utils.GetCDRClientOutOfCluster()
		svc := v1.Service{}
		crdClient.Get().AbsPath("/api/v1/namespaces/" + ns + "/services/" + funcName + "/").Do().Into(&svc)
		if svc.ObjectMeta.Name != funcName {
			logrus.Fatalf("Unable to find the service for %s", funcName)
		}

		port := strconv.Itoa(int(svc.Spec.Ports[0].Port))
		if svc.Spec.Ports[0].Name != "" {
			port = svc.Spec.Ports[0].Name
		}
		url := "/api/v1/proxy/namespaces/" + ns + "/services/" + funcName + ":" + port + "/"

		req := &rest.Request{}
		if get {
			req = crdClient.Get().AbsPath(url)
		} else {
			req = crdClient.Post().AbsPath(url).Body(bytes.NewBuffer(jsonStr)).SetHeader("Content-Type", "application/json")
		}
		res, err := req.Do().Raw()
		if err != nil {
			// Properly interpret line breaks
			logrus.Error(string(res))
			logrus.Fatal(strings.Replace(err.Error(), `\n`, "\n", -1))
		}
		fmt.Println(string(res))
	},
}

func init() {
	callCmd.Flags().StringP("data", "", "", "Specify data for function")
	callCmd.Flags().StringP("namespace", "", api.NamespaceDefault, "Specify namespace for the function")

}
