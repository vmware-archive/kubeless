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

package http

import (
	"github.com/kubeless/kubeless/pkg/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	kubelessApi "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var createCmd = &cobra.Command{
	Use:   "create <http_trigger_name> FLAG",
	Short: "Create a http trigger",
	Long:  `Create a http trigger`,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) != 1 {
			logrus.Fatal("Need exactly one argument - http trigger name")
		}
		triggerName := args[0]

		ns, err := cmd.Flags().GetString("namespace")
		if err != nil {
			logrus.Fatal(err)
		}
		if ns == "" {
			ns = utils.GetDefaultNamespace()
		}

		path, err := cmd.Flags().GetString("path")
		if err != nil {
			logrus.Fatal(err)
		}

		functionName, err := cmd.Flags().GetString("function-name")
		if err != nil {
			logrus.Fatal(err)
		}

		kubelessClient, err := utils.GetKubelessClientOutCluster()
		if err != nil {
			logrus.Fatalf("Can not create out-of-cluster client: %v", err)
		}

		_, err = utils.GetFunctionCustomResource(kubelessClient, functionName, ns)
		if err != nil {
			logrus.Fatalf("Unable to find Function %s in namespace %s. Error %s", functionName, ns, err)
		}

		httpTrigger := kubelessApi.HTTPTrigger{}
		httpTrigger.TypeMeta = metav1.TypeMeta{
			Kind:       "HTTPTrigger",
			APIVersion: "kubeless.io/v1beta1",
		}
		httpTrigger.ObjectMeta = metav1.ObjectMeta{
			Name:      triggerName,
			Namespace: ns,
		}
		httpTrigger.ObjectMeta.Labels = map[string]string{
			"created-by": "kubeless",
		}
		httpTrigger.Spec.FunctionName = functionName

		if len(path) != 0 {
			httpTrigger.Spec.Path = path
		}

		enableTLSAcme, err := cmd.Flags().GetBool("enableTLSAcme")
		if err != nil {
			logrus.Fatal(err)
		}
		httpTrigger.Spec.TLSAcme = enableTLSAcme

		tlsSecret, err := cmd.Flags().GetString("tls-secret")
		if err != nil {
			logrus.Fatal(err)
		}
		if enableTLSAcme && len(tlsSecret) > 0 {
			logrus.Fatalf("Cannot specify both --enableTLSAcme and --tls-secret")
		}
		httpTrigger.Spec.TLSSecret = tlsSecret

		gateway, err := cmd.Flags().GetString("gateway")
		if err != nil {
			logrus.Fatal(err)
		}
		if gateway != "nginx" && gateway != "traefik" && gateway != "kong" {
			logrus.Fatalf("Unsupported gateway %s", gateway)
		}
		httpTrigger.Spec.Gateway = gateway

		hostName, err := cmd.Flags().GetString("hostname")
		if err != nil {
			logrus.Fatal(err)
		}
		if hostName == "" && gateway == "nginx" {
			// We assume that Nginx will be listening in the port 80
			// of the cluster plublic IP
			config, err := utils.BuildOutOfClusterConfig()
			if err != nil {
				logrus.Fatal(err)
			}
			hostName, err = utils.GetLocalHostname(config, functionName)
			if err != nil {
				logrus.Fatal(err)
			}
		}
		if hostName == "" {
			logrus.Fatalf("The --hostname flag is required")
		}
		httpTrigger.Spec.HostName = hostName

		basicAuthSecret, err := cmd.Flags().GetString("basic-auth-secret")
		if err != nil {
			logrus.Fatal(err)
		}
		httpTrigger.Spec.BasicAuthSecret = basicAuthSecret

		err = utils.CreateHTTPTriggerCustomResource(kubelessClient, &httpTrigger)
		if err != nil {
			logrus.Fatalf("Failed to deploy HTTP trigger %s in namespace %s. Error: %s", triggerName, ns, err)
		}
		logrus.Infof("HTTP trigger %s created in namespace %s successfully!", triggerName, ns)
	},
}

func init() {
	createCmd.Flags().StringP("namespace", "", "", "Specify namespace for the HTTP trigger")
	createCmd.Flags().StringP("function-name", "", "", "Name of the function to be associated with trigger")
	createCmd.Flags().StringP("path", "", "", "Ingress path for the function")
	createCmd.Flags().StringP("hostname", "", "", "Specify a valid hostname for the function")
	createCmd.Flags().BoolP("enableTLSAcme", "", false, "If true, routing rule will be configured for use with kube-lego")
	createCmd.Flags().StringP("gateway", "", "nginx", "Specify a valid gateway for the Ingress. Supported: nginx, traefik, kong")
	createCmd.Flags().StringP("basic-auth-secret", "", "", "Specify an existing secret name for basic authentication")
	createCmd.Flags().StringP("tls-secret", "", "", "Specify an existing secret that contains a TLS private key and certificate to secure ingress")
	createCmd.MarkFlagRequired("function-name")
}
