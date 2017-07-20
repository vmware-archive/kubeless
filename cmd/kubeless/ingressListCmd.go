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
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/kubeless/kubeless/pkg/utils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"io"
	"k8s.io/client-go/pkg/api"

	"github.com/olekukonko/tablewriter"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ingressListCmd = &cobra.Command{
	Use:   "list FLAG",
	Short: "list all routes in Kubeless",
	Long:  `list all routes in Kubeless`,
	Run: func(cmd *cobra.Command, args []string) {
		output, err := cmd.Flags().GetString("out")
		if err != nil {
			logrus.Fatal(err.Error())
		}
		ns, err := cmd.Flags().GetString("namespace")
		if err != nil {
			logrus.Fatal(err.Error())
		}

		client := utils.GetClientOutOfCluster()

		if err := doIngressList(cmd.OutOrStdout(), client, ns, output); err != nil {
			logrus.Fatal(err.Error())
		}
	},
}

func init() {
	ingressListCmd.Flags().StringP("out", "o", "", "Output format. One of: json|yaml")
	ingressListCmd.Flags().StringP("namespace", "n", api.NamespaceDefault, "Specify namespace for the ingress")
}

func doIngressList(w io.Writer, client kubernetes.Interface, ns, output string) error {
	ingList, err := client.ExtensionsV1beta1().Ingresses(ns).List(v1.ListOptions{})
	if err != nil {
		return err
	}

	return prinIngress(w, ingList.Items, output)
}

// prinIngress formats the output of ingress list
func prinIngress(w io.Writer, ings []v1beta1.Ingress, output string) error {
	if output == "" {
		table := tablewriter.NewWriter(w)
		table.SetHeader([]string{"Name", "namespace", "host", "path", "service name", "service port"})
		for _, i := range ings {
			n := i.Name
			h := i.Spec.Rules[0].Host
			p := i.Spec.Rules[0].HTTP.Paths[0].Path
			ns := i.Namespace
			sN := i.Spec.Rules[0].HTTP.Paths[0].Backend.ServiceName
			sP := i.Spec.Rules[0].HTTP.Paths[0].Backend.ServicePort
			table.Append([]string{n, ns, h, p, sN, sP.String()})
		}
		table.Render()
	} else {
		for _, i := range ings {
			switch output {
			case "json":
				b, _ := json.MarshalIndent(i.Spec, "", "  ")
				fmt.Fprintln(w, string(b))
			case "yaml":
				b, _ := yaml.Marshal(i.Spec)
				fmt.Fprintln(w, string(b))
			default:
				return fmt.Errorf("Wrong output format. Please use only json|yaml")
			}
		}
	}
	return nil
}
