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
	"io"

	"github.com/Sirupsen/logrus"
	"github.com/gosuri/uitable"
	"github.com/kubeless/kubeless/pkg/utils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/apis/autoscaling/v2alpha1"
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
	asList, err := client.AutoscalingV2alpha1().HorizontalPodAutoscalers(ns).List(v1.ListOptions{})
	if err != nil {
		return err
	}

	return printAutoscale(w, asList.Items, output)
}

// printAutoscale formats the output of autoscale list
func printAutoscale(w io.Writer, ass []v2alpha1.HorizontalPodAutoscaler, output string) error {
	if output == "" {
		table := uitable.New()
		table.MaxColWidth = 50
		table.Wrap = true
		table.AddRow("NAME", "NAMESPACE", "TARGET", "MIN", "MAX", "TYPE", "OBJECT", "PODS", "RESOURCE")
		for _, i := range ass {
			n := i.Name
			ns := i.Namespace
			data, err := json.Marshal(i.Spec.ScaleTargetRef)
			if err != nil {
				return err
			}
			ta := string(data)
			min := string(*i.Spec.MinReplicas)
			max := string(i.Spec.MaxReplicas)
			ty := string(i.Spec.Metrics[0].Type)
			data, err = json.Marshal(i.Spec.Metrics[0].Object)
			if err != nil {
				return err
			}
			o := string(data)
			data, err = json.Marshal(i.Spec.Metrics[0].Pods)
			if err != nil {
				return err
			}
			p := string(data)
			data, err = json.Marshal(i.Spec.Metrics[0].Resource)
			if err != nil {
				return err
			}
			r := string(data)

			table.AddRow([]string{n, ns, ta, min, max, ty, o, p, r})
		}
		fmt.Fprintln(w, table)
	} else {
		for _, i := range ass {
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
