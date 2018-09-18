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
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/ghodss/yaml"
	"github.com/gosuri/uitable"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"

	kubelessApi "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	"github.com/kubeless/kubeless/pkg/client/clientset/versioned"
	"github.com/kubeless/kubeless/pkg/utils"
)

var topCmd = &cobra.Command{
	Use:     "top",
	Aliases: []string{"stats"},
	Short:   "display function metrics",
	Long:    `display function metrics`,
	Run: func(cmd *cobra.Command, args []string) {
		functionName, err := cmd.Flags().GetString("function")
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
		output, err := cmd.Flags().GetString("out")
		if err != nil {
			logrus.Fatal(err.Error())
		}

		apiV1Client := utils.GetClientOutOfCluster()
		kubelessClient, err := utils.GetKubelessClientOutCluster()
		handler := &utils.PrometheusMetricsHandler{}

		err = doTop(cmd.OutOrStdout(), kubelessClient, apiV1Client, handler, ns, functionName, output)
		if err != nil {
			logrus.Fatal(err.Error())
		}
	},
}

func init() {
	topCmd.Flags().StringP("namespace", "n", "", "Specify namespace for the function")
	topCmd.Flags().StringP("function", "f", "", "Specify the function")
	topCmd.Flags().StringP("out", "o", "", "Output format. One of: json|yaml")
}

func doTop(w io.Writer, kubelessClient versioned.Interface, apiV1Client kubernetes.Interface, handler utils.MetricsRetriever, ns, functionName, output string) error {
	functions, err := getFunctions(kubelessClient, ns, functionName)
	if err != nil {
		return fmt.Errorf("Error listing functions: %v", err)
	}

	ch := make(chan []*utils.Metric, len(functions))
	for _, f := range functions {
		go func(f *kubelessApi.Function) {
			ch <- utils.GetFunctionMetrics(apiV1Client, handler, ns, f.ObjectMeta.Name)
		}(f)
	}

	var metrics []*utils.Metric

	i := 0
	for i < len(functions) {
		select {
		case r := <-ch:
			metrics = append(metrics, r...)
			i++
		// timeout all go routines after 5 seconds to avoid hanging at the cmd line
		case <-time.After(5 * time.Second):
			i = len(functions)
		}
	}

	// sort the results - useful when using 'watch kubeless function top'
	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].FunctionName < metrics[j].FunctionName
	})
	return printTop(w, metrics, apiV1Client, output)
}

func printTop(w io.Writer, metrics []*utils.Metric, cli kubernetes.Interface, output string) error {
	if output == "" {
		table := uitable.New()
		table.MaxColWidth = 50
		table.Wrap = true
		table.AddRow("NAME", "NAMESPACE", "METHOD", "TOTAL_CALLS", "TOTAL_FAILURES", "TOTAL_DURATION_SECONDS", "AVG_DURATION_SECONDS", "MESSAGE")
		for _, f := range metrics {
			if f.Message != "" {
				table.AddRow(f.FunctionName, f.Namespace, "", "", "", "", "", f.Message)
			} else {
				table.AddRow(f.FunctionName, f.Namespace, f.Method, f.TotalCalls, f.TotalFailures, f.TotalDurationSeconds, f.AvgDurationSeconds, "")
			}
		}
		fmt.Fprintln(w, table)
	} else {
		switch output {
		case "json":
			b, err := json.MarshalIndent(metrics, "", "  ")
			if err != nil {
				return err
			}
			fmt.Fprintln(w, string(b))
		case "yaml":
			b, err := yaml.Marshal(metrics)
			if err != nil {
				return err
			}
			fmt.Fprintln(w, string(b))
		default:
			return fmt.Errorf("Wrong output format. Please use only json|yaml")
		}
	}
	return nil
}
