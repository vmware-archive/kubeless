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
	"github.com/Sirupsen/logrus"
	"github.com/kubeless/kubeless/pkg/spec"
	"github.com/kubeless/kubeless/pkg/utils"
	"github.com/spf13/cobra"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/rest"
)

var ingressCreateCmd = &cobra.Command{
	Use:   "create <name> FLAG",
	Short: "create a route to function",
	Long:  `create a route to function`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			logrus.Fatal("Need exactly one argument - ingress name")
		}
		ingressName := args[0]
		ns, err := cmd.Flags().GetString("namespace")
		if err != nil {
			logrus.Fatal(err)
		}

		funcName, err := cmd.Flags().GetString("function")
		if err != nil {
			logrus.Fatal(err)
		}

		hostName, err := cmd.Flags().GetString("hostname")
		if err != nil {
			logrus.Fatal(err)
		}
		if hostName == "" {
			config, err := utils.BuildOutOfClusterConfig()
			if err != nil {
				logrus.Fatal(err)
			}
			hostName, err = utils.GetLocalHostname(config, funcName)
			if err != nil {
				logrus.Fatal(err)
			}
		}

		tprClient, err := utils.GetTPRClientOutOfCluster()
		if err != nil {
			logrus.Fatal(err)
		}
		err = functionExists(tprClient, funcName, ns)
		if err != nil {
			if k8sErrors.IsNotFound(err) {
				logrus.Fatalf("function %s doesn't exist in namespace %s", funcName, ns)
			} else {
				logrus.Fatalf("error validate input %v", err)
			}
		}

		client := utils.GetClientOutOfCluster()

		err = utils.CreateIngress(client, ingressName, funcName, hostName, ns)
		if err != nil {
			logrus.Fatalf("Can't create ingress route: %v", err)
		}
	},
}

func functionExists(tprClient rest.Interface, function, ns string) error {
	f := spec.Function{}
	err := tprClient.Get().
		Resource("functions").
		Namespace(ns).
		Name(function).
		Do().
		Into(&f)
	return err
}

func init() {
	ingressCreateCmd.Flags().StringP("hostname", "", "", "Specify a valid hostname for the function")
	ingressCreateCmd.Flags().StringP("function", "", "", "Name of the function")
}
