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

package utils

import (
	"bytes"

	"k8s.io/client-go/kubernetes"

	"github.com/prometheus/common/expfmt"
)

// Metric contains metrics for a functions
type Metric struct {
	FunctionName         string  `json:"function,omitempty"`
	Namespace            string  `json:"namespace,omitempty"`
	Method               string  `json:"method,omitempty"`
	Message              string  `json:"message,omitempty"`
	TotalCalls           float64 `json:"total_calls,omitempty"`
	TotalFailures        float64 `json:"total_failures,omitempty"`
	TotalDurationSeconds float64 `json:"total_duration_seconds,omitempty"`
	AvgDurationSeconds   float64 `json:"avg_duration_seconds,omitempty"`
}

// MetricsRetriever is an interface for retreiving metrics from an endpoint
type MetricsRetriever interface {
	GetRawMetrics(kubernetes.Interface, string, string) ([]byte, error)
}

// PrometheusMetricsHandler is a handler for retreiving metrics from Prometheus
type PrometheusMetricsHandler struct{}

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

// GetRawMetrics returns the raw metrics for a Prometheus endpoint
func (h *PrometheusMetricsHandler) GetRawMetrics(apiV1Client kubernetes.Interface, namespace, functionName string) ([]byte, error) {

	port, err := GetFunctionPort(apiV1Client, namespace, functionName)
	if err != nil {
		return []byte{}, err
	}

	req := apiV1Client.CoreV1().RESTClient().Get().Namespace(namespace).Resource("services").SubResource("proxy").Name(functionName + ":" + port).Suffix("/metrics")
	return req.Do().Raw()
}

// GetFunctionMetrics returns Prometheus metrics as a slice of *Metrics
func GetFunctionMetrics(apiV1Client kubernetes.Interface, h MetricsRetriever, namespace, functionName string) []*Metric {

	res, err := h.GetRawMetrics(apiV1Client, namespace, functionName)
	if err != nil {
		return []*Metric{
			{
				FunctionName: functionName,
				Namespace:    namespace,
				Message:      "Function does not expose metrics",
			},
		}
	}

	metrics, err := parseMetrics(namespace, functionName, res)
	if err != nil {
		return []*Metric{
			{
				FunctionName: functionName,
				Namespace:    namespace,
				Message:      "Unable to get function metrics",
			},
		}
	}
	return metrics
}
