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

package function

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/kubeless/kubeless/pkg/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

var callCmd = &cobra.Command{
	Use:   "call <function_name> FLAG",
	Short: "call function from cli",
	Long:  `call function from cli`,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			str []byte
			get bool = false
		)

		if len(args) != 1 {
			logrus.Fatal("Need exactly one argument - function name")
		}
		funcName := args[0]

		data, err := cmd.Flags().GetString("data")
		if data == "" {
			get = true
		} else {
			str = []byte(data)
		}

		if err != nil {
			logrus.Fatal(err)
		}
		ns, err := cmd.Flags().GetString("namespace")
		if err != nil {
			logrus.Fatal(err)
		}
		if ns == "" {
			ns = utils.GetDefaultNamespace()
		}

		clientset := utils.GetClientOutOfCluster()
		svc, err := clientset.CoreV1().Services(ns).Get(funcName, metav1.GetOptions{})
		if err != nil {
			logrus.Fatalf("Unable to find the service for %s", funcName)
		}

		port := strconv.Itoa(int(svc.Spec.Ports[0].Port))
		if svc.Spec.Ports[0].Name != "" {
			port = svc.Spec.Ports[0].Name
		}

		req := &rest.Request{}
		if get {
			req = clientset.CoreV1().RESTClient().Get().Namespace(ns).Resource("services").SubResource("proxy").Name(funcName + ":" + port)
		} else {
			req = clientset.CoreV1().RESTClient().Post().Body(bytes.NewBuffer(str))
			if utils.IsJSON(string(str)) {
				req.SetHeader("Content-Type", "application/json")
				req.SetHeader("event-type", "application/json")
			} else {
				req.SetHeader("Content-Type", "application/x-www-form-urlencoded")
				req.SetHeader("event-type", "application/x-www-form-urlencoded")
			}
			// REST package removes trailing slash when building URLs
			// Causing POST requests to be redirected with an empty body
			// So we need to manually build the URL
			req = req.AbsPath(svc.ObjectMeta.SelfLink + ":" + port + "/proxy/")
		}
		timestamp := time.Now().UTC()
		eventID, err := utils.GetRandString(11)
		if err != nil {
			logrus.Fatalf("Unable to generate ID %v", err)
		}
		req.SetHeader("event-id", eventID)
		req.SetHeader("event-time", timestamp.Format(time.RFC3339))
		req.SetHeader("event-namespace", "cli.kubeless.io")
		res, err := req.Do().Raw()
		if err != nil {
			// Properly interpret line breaks
			logrus.Error(string(res))
			if strings.Contains(err.Error(), "status code 408") {
				// Give a more meaninful error for timeout errors
				logrus.Fatal("Request timeout exceeded")
			} else {
				logrus.Fatal(strings.Replace(err.Error(), `\n`, "\n", -1))
			}
		}
		fmt.Println(string(res))
	},
}

func init() {
	callCmd.Flags().StringP("data", "d", "", "Specify data for function")
	callCmd.Flags().StringP("namespace", "n", "", "Specify namespace for the function")

}
