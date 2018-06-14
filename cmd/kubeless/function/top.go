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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
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
		handler := &PrometheusMetricsHandler{}

		err = doTop(cmd.OutOrStdout(), kubelessClient, apiV1Client, handler, ns, functionName, output)
		if err != nil {
			logrus.Fatal(err.Error())
		}
	},
}

// Metric contains metrics for a functions
type Metric struct {
	FunctionName         string  `json:"function,omitempty"`
	Namespace            string  `json:"namespace,omitempty"`
	Method               string  `json:"method,omitempty"`
	TotalCalls           float64 `json:"total_calls,omitempty"`
	TotalFailures        float64 `json:"total_failures,omitempty"`
	TotalDurationSeconds float64 `json:"total_duration_seconds,omitempty"`
	AvgDurationSeconds   float64 `json:"avg_duration_seconds,omitempty"`
}

func init() {
	topCmd.Flags().StringP("namespace", "n", "", "Specify namespace for the function")
	topCmd.Flags().StringP("function", "f", "", "Specify the function")
	topCmd.Flags().StringP("out", "o", "", "Output format. One of: json|yaml")
}

// MetricsRetriever is an interface for retreiving metrics from an endpoint
type MetricsRetriever interface {
	getRawMetrics(kubernetes.Interface, string, string) ([]byte, error)
}

// PrometheusMetricsHandler is a handler for retreiving metrics from Prometheus
type PrometheusMetricsHandler struct{}

func getFunctions(kubelessClient versioned.Interface, namespace, functionName string) ([]*kubelessApi.Function, error) {
	var functionList *kubelessApi.FunctionList
	var f *kubelessApi.Function
	var err error
	if functionName == "" {
		functionList, err = kubelessClient.KubelessV1beta1().Functions(namespace).List(metav1.ListOptions{})
		if err != nil {
			return []*kubelessApi.Function{}, err
		}
	} else {
		f, err = kubelessClient.KubelessV1beta1().Functions(namespace).Get(functionName, metav1.GetOptions{})
		if err != nil {
			return []*kubelessApi.Function{}, err
		}
		functionList = &kubelessApi.FunctionList{
			Items: []*kubelessApi.Function{
				f,
			},
		}
	}
	return functionList.Items, nil
}

func parseMetrics(namespace, functionName string, rawMetrics []byte) ([]*Metric, error) {
	parser := expfmt.TextParser{}
	parsedData, err := parser.TextToMetricFamilies(bytes.NewReader(rawMetrics))
	if err != nil {
		return nil, err
	}

	tmp := map[string]*Metric{}
	var parsedMetrics []*Metric

	metricsOfInterest := []string{"function_duration_seconds", "function_calls_total", "function_failures_total"}
	for _, m := range metricsOfInterest {
		for _, metric := range parsedData[m].GetMetric() {
			// a function can have metrics for multiple methods (GET, POST, etc.)
			// method names can be values other than GET/POST/PUT/DELETE
			for _, label := range metric.GetLabel() {
				if label.GetName() == "method" {
					if _, ok := tmp[label.GetValue()]; !ok {
						tmp[label.GetValue()] = &Metric{
							FunctionName: functionName,
							Namespace:    namespace,
							Method:       label.GetValue(),
						}
					}
					if m == "function_failures_total" {
						tmp[label.GetValue()].TotalFailures = metric.GetCounter().GetValue()
					}
					if m == "function_duration_seconds" {
						tmp[label.GetValue()].TotalDurationSeconds = metric.GetHistogram().GetSampleSum()
					}
					if m == "function_calls_total" {
						tmp[label.GetValue()].TotalCalls = metric.GetCounter().GetValue()
						if tmp[label.GetValue()].TotalCalls > 0 {
							tmp[label.GetValue()].AvgDurationSeconds = float64(tmp[label.GetValue()].TotalDurationSeconds) / tmp[label.GetValue()].TotalCalls
						}
					}
				}
			}
		}
	}

	// if the funciton hasn't been invoked, add an item to the list so the function displays in the output
	if len(tmp) == 0 {
		tmp[""] = &Metric{
			FunctionName: functionName,
			Namespace:    namespace,
		}
	}

	for _, v := range tmp {
		parsedMetrics = append(parsedMetrics, v)
	}

	return parsedMetrics, nil
}

func (h *PrometheusMetricsHandler) getRawMetrics(apiV1Client kubernetes.Interface, namespace, functionName string) ([]byte, error) {
	svc, err := apiV1Client.CoreV1().Services(namespace).Get(functionName, metav1.GetOptions{})
	if err != nil {
		return []byte{}, err
	}

	port := strconv.Itoa(int(svc.Spec.Ports[0].Port))
	if svc.Spec.Ports[0].Name != "" {
		port = svc.Spec.Ports[0].Name
	}
	req := apiV1Client.CoreV1().RESTClient().Get().Namespace(namespace).Resource("services").SubResource("proxy").Name(functionName + ":" + port).Suffix("/metrics")
	return req.Do().Raw()
}

func getFunctionMetrics(apiV1Client kubernetes.Interface, h MetricsRetriever, namespace, functionName string) []*Metric {

	res, err := h.getRawMetrics(apiV1Client, namespace, functionName)
	if err != nil {
		return []*Metric{}
	}

	metrics, err := parseMetrics(namespace, functionName, res)
	if err != nil {
		return []*Metric{}
	}
	return metrics
}

func doTop(w io.Writer, kubelessClient versioned.Interface, apiV1Client kubernetes.Interface, handler MetricsRetriever, ns, functionName, output string) error {
	functions, err := getFunctions(kubelessClient, ns, functionName)
	if err != nil {
		return fmt.Errorf("Error listing functions: %v", err)
	}

	ch := make(chan []*Metric, len(functions))
	for _, f := range functions {
		go func(f *kubelessApi.Function) {
			ch <- getFunctionMetrics(apiV1Client, handler, ns, f.ObjectMeta.Name)
		}(f)
	}

	var metrics []*Metric

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

func printTop(w io.Writer, metrics []*Metric, cli kubernetes.Interface, output string) error {
	if output == "" {
		table := uitable.New()
		table.MaxColWidth = 50
		table.Wrap = true
		table.AddRow("NAME", "NAMESPACE", "METHOD", "TOTAL_CALLS", "TOTAL_FAILURES", "TOTAL_DURATION_SECONDS", "AVG_DURATION_SECONDS")
		for _, f := range metrics {
			table.AddRow(f.FunctionName, f.Namespace, f.Method, f.TotalCalls, f.TotalFailures, f.TotalDurationSeconds, f.AvgDurationSeconds)
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
