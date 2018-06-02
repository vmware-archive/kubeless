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
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gosuri/uitable"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	kubelessApi "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	"github.com/kubeless/kubeless/pkg/client/clientset/versioned"
	"github.com/kubeless/kubeless/pkg/utils"

	"github.com/prometheus/common/expfmt"
)

// TODO - testing
// TODO - wide view
// TODO - add timing information to the display - how long it took to query/parse the stats

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

		apiV1Client := utils.GetClientOutOfCluster()
		kubelessClient, err := utils.GetKubelessClientOutCluster()

		err = doTop(cmd.OutOrStdout(), kubelessClient, apiV1Client, ns, functionName, args)
		if err != nil {
			logrus.Fatal(err.Error())
		}
	},
}

type functionMetrics struct {
	functionName       string
	namespace          string
	functionRawMetrics string
	// map[METHOD][METRIC_NAME]
	metricsMap map[string]map[string]float64
}

func init() {
	topCmd.Flags().StringP("namespace", "n", "", "Specify namespace for the function")
	topCmd.Flags().StringP("function", "f", "", "Specify the function")
}

func getFunctions(kubelessClient versioned.Interface, namespace, functionName string) ([]*kubelessApi.Function, error) {
	var functionList *kubelessApi.FunctionList
	var f *kubelessApi.Function
	var err error
	if functionName == "" {
		functionList, err = kubelessClient.KubelessV1beta1().Functions(namespace).List(metav1.ListOptions{})
	} else {
		f, err = kubelessClient.KubelessV1beta1().Functions(namespace).Get(functionName, metav1.GetOptions{})
		functionList = &kubelessApi.FunctionList{
			Items: []*kubelessApi.Function{
				f,
			},
		}
	}
	if err != nil {
		return []*kubelessApi.Function{}, err
	}
	return functionList.Items, nil
}

func parseMetrics(metrics *functionMetrics) error {
	parser := expfmt.TextParser{}
	parsedData, err := parser.TextToMetricFamilies(strings.NewReader(metrics.functionRawMetrics))
	if err != nil {
		return err
	}

	metricsOfInterest := []string{"function_duration_seconds", "function_calls_total", "function_failures_total"}
	for _, m := range metricsOfInterest {
		for _, metric := range parsedData[m].GetMetric() {
			for _, label := range metric.GetLabel() {
				if label.GetName() == "method" {
					if metrics.metricsMap[label.GetValue()] == nil {
						metrics.metricsMap[label.GetValue()] = make(map[string]float64)
					}
					if m == "function_duration_seconds" {
						metrics.metricsMap[label.GetValue()][m] = metric.GetHistogram().GetSampleSum()
					} else {
						metrics.metricsMap[label.GetValue()][m] = metric.GetCounter().GetValue()
					}
				}
			}
		}
	}
	return nil
}

func calculateMetrics(metrics *functionMetrics) {
	for _, v := range metrics.metricsMap {
		if v["function_calls_total"] > 0 {
			v["function_average_duration_seconds"] = v["function_duration_seconds"] / v["function_calls_total"]
		}
	}
}

func getFunctionMetrics(apiV1Client kubernetes.Interface, namespace, functionName string, responseChan chan functionMetrics) {
	metrics := functionMetrics{}
	metrics.functionName = functionName
	metrics.namespace = namespace
	metrics.metricsMap = make(map[string]map[string]float64)
	svc, err := apiV1Client.CoreV1().Services(namespace).Get(functionName, metav1.GetOptions{})
	if err != nil {
		responseChan <- metrics
		return
	}
	port := strconv.Itoa(int(svc.Spec.Ports[0].Port))
	req := apiV1Client.CoreV1().RESTClient().Get().Namespace(namespace).Resource("services").SubResource("proxy").Name(functionName + ":" + port).Suffix("/metrics")
	res, err := req.Do().Raw()
	if err != nil {
		responseChan <- metrics
		return
	}
	metrics.functionRawMetrics = string(res)

	_ = parseMetrics(&metrics)
	calculateMetrics(&metrics)
	responseChan <- metrics
}

func doTop(w io.Writer, kubelessClient versioned.Interface, apiV1Client kubernetes.Interface, ns, functionName string, args []string) error {
	functions, err := getFunctions(kubelessClient, ns, functionName)
	if err != nil {
		return fmt.Errorf("Error listing functions: %v", err)
	}

	var metrics []functionMetrics
	ch := make(chan functionMetrics, len(functions))
	for _, f := range functions {
		go getFunctionMetrics(apiV1Client, ns, f.ObjectMeta.Name, ch)
	}

loop:
	for len(metrics) < len(functions) {
		select {
		case r := <-ch:
			metrics = append(metrics, r)
		case <-time.After(5 * time.Second):
			break loop
		}
	}

	sort.Slice(metrics, func(i, j int) bool { return metrics[i].functionName < metrics[j].functionName })
	return printTop(w, metrics, apiV1Client)
}

func printTop(w io.Writer, metrics []functionMetrics, cli kubernetes.Interface) error {
	table := uitable.New()
	table.MaxColWidth = 50
	table.Wrap = true
	table.AddRow("NAME", "NAMESPACE", "METHOD", "TOTAL_CALLS", "TOTAL_FAILURES", "DURATION_SECONDS", "AVG_DURATION_SECONDS")
	for _, f := range metrics {
		n := f.functionName
		ns := f.namespace
		if len(f.metricsMap) > 0 {
			for k, metrics := range f.metricsMap {
				method := k
				totalCalls := strconv.FormatFloat(metrics["function_calls_total"], 'f', -1, 32)
				durationSeconds := strconv.FormatFloat(metrics["function_duration_seconds"], 'f', -1, 32)
				totalFailures := strconv.FormatFloat(metrics["function_failures_total"], 'f', -1, 32)
				avgDuration := strconv.FormatFloat(metrics["function_average_duration_seconds"], 'f', -1, 32)
				table.AddRow(n, ns, method, totalCalls, totalFailures, durationSeconds, avgDuration)
			}
		} else {
			table.AddRow(n, ns)
		}
	}
	fmt.Fprintln(w, table)
	return nil
}
