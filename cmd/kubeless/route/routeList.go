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

package route

import (
	"encoding/json"
	"fmt"
	"github.com/kubeless/kubeless/pkg/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"io"

	"github.com/gosuri/uitable"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

var routeListCmd = &cobra.Command{
	Use:     "list FLAG",
	Aliases: []string{"ls"},
	Short:   "list all routes in Kubeless",
	Long:    `list all routes in Kubeless`,
	Run: func(cmd *cobra.Command, args []string) {
		output, err := cmd.Flags().GetString("out")
		if err != nil {
			logrus.Fatal(err.Error())
		}
		ns, err := cmd.Flags().GetString("namespace")
		if err != nil {
			logrus.Fatal(err.Error())
		}
		if ns == "" {
			ns = utils.GetDefaultNamespace()
		}

		client := utils.GetClientOutOfCluster()

		if err := doIngressList(cmd.OutOrStdout(), client, ns, output); err != nil {
			logrus.Fatal(err.Error())
		}
	},
}

func init() {
	routeListCmd.Flags().StringP("out", "o", "", "Output format. One of: json|yaml")
}

func doIngressList(w io.Writer, client kubernetes.Interface, ns, output string) error {
	ingList, err := client.ExtensionsV1beta1().Ingresses(ns).List(v1.ListOptions{
		LabelSelector: "created-by=kubeless",
	})
	if err != nil {
		return err
	}

	return printIngress(w, ingList.Items, output)
}

// printIngress formats the output of ingress list
func printIngress(w io.Writer, ings []v1beta1.Ingress, output string) error {
	if output == "" {
		table := uitable.New()
		table.MaxColWidth = 50
		table.Wrap = true
		table.AddRow("NAME", "NAMESPACE", "HOST", "PATH", "SERVICE NAME", "SERVICE PORT")
		for _, i := range ings {
			if len(i.Spec.Rules) == 0 {
				logrus.Errorf("The function route %s isn't in correct format. It has no rule defined.", i.Name)
				continue
			}
			n := i.Name
			h := i.Spec.Rules[0].Host
			p := i.Spec.Rules[0].HTTP.Paths[0].Path
			ns := i.Namespace
			sN := i.Spec.Rules[0].HTTP.Paths[0].Backend.ServiceName
			sP := i.Spec.Rules[0].HTTP.Paths[0].Backend.ServicePort
			table.AddRow(n, ns, h, p, sN, sP.String())
		}
		fmt.Fprintln(w, table)
	} else {
		for _, i := range ings {
			switch output {
			case "json":
				b, err := json.MarshalIndent(i.Spec, "", "  ")
				if err != nil {
					return err
				}
				fmt.Fprintln(w, string(b))
			case "yaml":
				b, err := yaml.Marshal(i.Spec)
				if err != nil {
					return err
				}
				fmt.Fprintln(w, string(b))
			default:
				return fmt.Errorf("Wrong output format. Please use only json|yaml")
			}
		}
	}
	return nil
}
