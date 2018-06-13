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
	"strconv"
	"strings"
	"time"

	"github.com/ghodss/yaml"
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
// TODO - yaml, json

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
		handler := &PrometheusMetricsHandler{metricsPath: "/metrics"}

		err = doTop(cmd.OutOrStdout(), kubelessClient, apiV1Client, handler, ns, functionName, output)
		if err != nil {
			logrus.Fatal(err.Error())
		}
	},
}

type functionMetrics struct {
	FunctionName       string `json:"function_name,omitempty"`
	Namespace          string `json:"namespace,omitempty"`
	FunctionRawMetrics string `json:"function_raw_metrics,omitempty"`
	// map[METHOD][METRIC_NAME]
	MetricsMap map[string]map[string]float64 `json:",omitempty"`
}

func init() {
	topCmd.Flags().StringP("namespace", "n", "", "Specify namespace for the function")
	topCmd.Flags().StringP("function", "f", "", "Specify the function")
	topCmd.Flags().StringP("out", "o", "", "Output format. One of: json|yaml")
}

// MetricsRetriever is an interface for retreiving metrics from an endpoint
type MetricsRetriever interface {
	getMetrics(kubernetes.Interface, string, string, string) ([]byte, error)
}

// PrometheusMetricsHandler is a handler for retreiving metrics from Prometheus
type PrometheusMetricsHandler struct {
	metricsPath string
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
	parsedData, err := parser.TextToMetricFamilies(strings.NewReader(metrics.FunctionRawMetrics))
	if err != nil {
		fmt.Printf("error while parsing metrics ", err)
		return err
	}

	metricsOfInterest := []string{"function_duration_seconds", "function_calls_total", "function_failures_total"}
	for _, m := range metricsOfInterest {
		for _, metric := range parsedData[m].GetMetric() {
			// a function can have metrics for each method (GET, POST, etc.) - group metrics by method and show a single line per method
			for _, label := range metric.GetLabel() {
				if label.GetName() == "method" {
					if metrics.MetricsMap[label.GetValue()] == nil {
						metrics.MetricsMap[label.GetValue()] = make(map[string]float64)
					}
					if m == "function_duration_seconds" {
						metrics.MetricsMap[label.GetValue()][m] = metric.GetHistogram().GetSampleSum()
					} else {
						metrics.MetricsMap[label.GetValue()][m] = metric.GetCounter().GetValue()
					}
				}
			}
		}
	}
	return nil
}

func calculateMetrics(metrics *functionMetrics) {
	for _, v := range metrics.MetricsMap {
		if v["function_calls_total"] > 0 {
			v["function_average_duration_seconds"] = v["function_duration_seconds"] / v["function_calls_total"]
		}
	}
}

func (h *PrometheusMetricsHandler) getMetrics(apiV1Client kubernetes.Interface, namespace, functionName, port string) ([]byte, error) {
	req := apiV1Client.CoreV1().RESTClient().Get().Namespace(namespace).Resource("services").SubResource("proxy").Name(functionName + ":" + port).Suffix(h.metricsPath)
	res, err := req.Do().Raw()
	if err != nil {
		return []byte{}, err
	}
	return res, nil
}

func getFunctionMetrics(apiV1Client kubernetes.Interface, h MetricsRetriever, namespace, functionName string) functionMetrics {
	metrics := functionMetrics{}
	metrics.FunctionName = functionName
	metrics.Namespace = namespace
	metrics.MetricsMap = make(map[string]map[string]float64)
	svc, err := apiV1Client.CoreV1().Services(namespace).Get(functionName, metav1.GetOptions{})
	if err != nil {
		return metrics
	}

	port := strconv.Itoa(int(svc.Spec.Ports[0].Port))
	if svc.Spec.Ports[0].Name != "" {
		port = svc.Spec.Ports[0].Name
	}
	res, err := h.getMetrics(apiV1Client, namespace, functionName, port)

	if err != nil {
		return metrics
	}
	metrics.FunctionRawMetrics = string(res)

	// swallow the error if metrics parsing fails - display will show the function name but no metrics
	_ = parseMetrics(&metrics)
	calculateMetrics(&metrics)
	return metrics
}

func doTop(w io.Writer, kubelessClient versioned.Interface, apiV1Client kubernetes.Interface, handler MetricsRetriever, ns, functionName, output string) error {
	functions, err := getFunctions(kubelessClient, ns, functionName)
	if err != nil {
		return fmt.Errorf("Error listing functions: %v", err)
	}

	var metrics []functionMetrics
	ch := make(chan functionMetrics, len(functions))
	for _, f := range functions {
		go func(f *kubelessApi.Function) {
			ch <- getFunctionMetrics(apiV1Client, handler, ns, f.ObjectMeta.Name)
		}(f)
	}

metricsLoop:
	for len(metrics) < len(functions) {
		select {
		case r := <-ch:
			metrics = append(metrics, r)
		// timeout all go routines after 5 seconds to avoid hanging at the cmd line
		case <-time.After(5 * time.Second):
			break metricsLoop
		}
	}

	// sort the results - useful when using 'watch kubeless function top'
	sort.Slice(metrics, func(i, j int) bool { return metrics[i].FunctionName < metrics[j].FunctionName })
	return printTop(w, metrics, apiV1Client, output)
}

func printTop(w io.Writer, metrics []functionMetrics, cli kubernetes.Interface, output string) error {
	if output == "" {
		table := uitable.New()
		table.MaxColWidth = 50
		table.Wrap = true
		table.AddRow("NAME", "NAMESPACE", "METHOD", "TOTAL_CALLS", "TOTAL_FAILURES", "TOTAL_DURATION_SECONDS", "AVG_DURATION_SECONDS")
		for _, f := range metrics {
			n := f.FunctionName
			ns := f.Namespace
			if len(f.MetricsMap) > 0 {
				for k, metrics := range f.MetricsMap {
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
