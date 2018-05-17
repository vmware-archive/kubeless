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

package autoscale

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/ghodss/yaml"
	"github.com/gosuri/uitable"
	"github.com/kubeless/kubeless/pkg/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/api/autoscaling/v2beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var autoscaleListCmd = &cobra.Command{
	Use:     "list FLAG",
	Aliases: []string{"ls"},
	Short:   "list all autoscales in Kubeless",
	Long:    `list all autoscales in Kubeless`,
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

		if err := doAutoscaleList(cmd.OutOrStdout(), client, ns, output); err != nil {
			logrus.Fatal(err.Error())
		}
	},
}

func init() {
	autoscaleListCmd.Flags().StringP("out", "o", "", "Output format. One of: json|yaml")
}

func doAutoscaleList(w io.Writer, client kubernetes.Interface, ns, output string) error {
	asList, err := client.AutoscalingV2beta1().HorizontalPodAutoscalers(ns).List(metav1.ListOptions{
		LabelSelector: "created-by=kubeless",
	})
	if err != nil {
		return err
	}

	return printAutoscale(w, asList.Items, output)
}

// printAutoscale formats the output of autoscale list
func printAutoscale(w io.Writer, ass []v2beta1.HorizontalPodAutoscaler, output string) error {
	if output == "" {
		table := uitable.New()
		table.MaxColWidth = 50
		table.Wrap = true
		table.AddRow("NAME", "NAMESPACE", "TARGET", "MIN", "MAX", "METRIC", "VALUE")
		for _, i := range ass {
			n := i.Name
			ns := i.Namespace
			ta := i.Spec.ScaleTargetRef.Name
			min := i.Spec.MinReplicas
			max := i.Spec.MaxReplicas
			m := ""
			v := ""
			if len(i.Spec.Metrics) == 0 {
				logrus.Errorf("The autoscale %s has bad format. It has no metric defined.", i.Name)
				continue
			}
			if i.Spec.Metrics[0].Object != nil {
				m = i.Spec.Metrics[0].Object.MetricName
				v = i.Spec.Metrics[0].Object.TargetValue.String()
			} else if i.Spec.Metrics[0].Resource != nil {
				m = string(i.Spec.Metrics[0].Resource.Name)
				v = fmt.Sprint(*i.Spec.Metrics[0].Resource.TargetAverageUtilization)
			}

			table.AddRow(n, ns, ta, fmt.Sprint(*min), fmt.Sprint(max), m, v)
		}
		fmt.Fprintln(w, table)
	} else {
		for _, i := range ass {
			switch output {
			case "json":
				b, err := json.MarshalIndent(i, "", "  ")
				if err != nil {
					return err
				}
				fmt.Fprintln(w, string(b))
			case "yaml":
				b, err := yaml.Marshal(i)
				if err != nil {
					return err
				}
				fmt.Fprintln(w, string(b))
			default:
				return fmt.Errorf("Wrong output format. Only accept json|yaml file")
			}
		}
	}
	return nil
}
