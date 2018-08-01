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
	"fmt"

	httpUtils "github.com/kubeless/http-trigger/pkg/utils"
	kubelessUtils "github.com/kubeless/kubeless/pkg/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update <http_trigger_name> FLAG",
	Short: "Update a http trigger",
	Long:  `Update a http trigger`,
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
			ns = kubelessUtils.GetDefaultNamespace()
		}

		kubelessClient, err := kubelessUtils.GetKubelessClientOutCluster()
		if err != nil {
			logrus.Fatalf("Can not create out-of-cluster client: %v", err)
		}

		httpClient, err := httpUtils.GetKubelessClientOutCluster()
		if err != nil {
			logrus.Fatalf("Can not create out-of-cluster client: %v", err)
		}

		httpTrigger, err := httpUtils.GetHTTPTriggerCustomResource(httpClient, triggerName, ns)
		if err != nil {
			logrus.Fatalf("Unable to find HTTP trigger %s in namespace %s. Error %s", triggerName, ns, err)
		}

		functionName, err := cmd.Flags().GetString("function-name")
		if err != nil {
			logrus.Fatal(err)
		}

		if functionName != "" {
			_, err = kubelessUtils.GetFunctionCustomResource(kubelessClient, functionName, ns)
			if err != nil {
				logrus.Fatalf("Unable to find Function %s in namespace %s. Error %s", functionName, ns, err)
			}
			httpTrigger.Spec.FunctionName = functionName
		}

		path, err := cmd.Flags().GetString("path")
		if err != nil {
			logrus.Fatal(err)
		}
		if path != "" {
			httpTrigger.Spec.Path = path
		}

		hostName, err := cmd.Flags().GetString("hostname")
		if err != nil {
			logrus.Fatal(err)
		}
		if hostName != "" {
			httpTrigger.Spec.HostName = hostName
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
		if tlsSecret != "" {
			httpTrigger.Spec.TLSSecret = tlsSecret
		}

		gateway, err := cmd.Flags().GetString("gateway")
		if err != nil {
			logrus.Fatal(err)
		}
		if gateway != "" {
			httpTrigger.Spec.Gateway = gateway
		}

		basicAuthSecret, err := cmd.Flags().GetString("basic-auth-secret")
		if err != nil {
			logrus.Fatal(err)
		}
		if basicAuthSecret != "" {
			httpTrigger.Spec.BasicAuthSecret = basicAuthSecret
		}

		dryrun, err := cmd.Flags().GetBool("dryrun")
		if err != nil {
			logrus.Fatal(err)
		}

		output, err := cmd.Flags().GetString("output")
		if err != nil {
			logrus.Fatal(err)
		}

		if dryrun == true {
			res, err := kubelessUtils.DryRunFmt(output, httpTrigger)
			if err != nil {
				logrus.Fatal(err)
			}
			fmt.Println(res)
			return
		}

		err = httpUtils.UpdateHTTPTriggerCustomResource(httpClient, httpTrigger)
		if err != nil {
			logrus.Fatalf("Failed to deploy HTTP trigger %s in namespace %s. Error: %s", triggerName, ns, err)
		}
		logrus.Infof("HTTP trigger %s updated in namespace %s successfully!", triggerName, ns)
	},
}

func init() {
	updateCmd.Flags().StringP("namespace", "n", "", "Specify namespace for the HTTP trigger")
	updateCmd.Flags().StringP("function-name", "", "", "Name of the function to be associated with trigger")
	updateCmd.Flags().StringP("path", "", "", "Ingress path for the function")
	updateCmd.Flags().StringP("hostname", "", "", "Specify a valid hostname for the function")
	updateCmd.Flags().BoolP("enableTLSAcme", "", false, "If true, routing rule will be configured for use with kube-lego")
	updateCmd.Flags().StringP("gateway", "", "", "Specify a valid gateway for the Ingress")
	updateCmd.Flags().StringP("basic-auth-secret", "", "", "Specify an existing secret name for basic authentication")
	updateCmd.Flags().StringP("tls-secret", "", "", "Specify an existing secret that contains a TLS private key and certificate to secure ingress")
	updateCmd.Flags().Bool("dryrun", false, "Output JSON manifest of the function without creating it")
	updateCmd.Flags().StringP("output", "o", "yaml", "Output format")
}
