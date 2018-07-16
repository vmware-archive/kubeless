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
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"

	kubelessApi "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	"github.com/kubeless/kubeless/pkg/client/clientset/versioned"
	fFake "github.com/kubeless/kubeless/pkg/client/clientset/versioned/fake"
	"github.com/kubeless/kubeless/pkg/utils"
)

type testMetricsHandler struct{}

// handler used for testing purposes only
// satisfies the MetricsRetriever interface, gets metrics from the test http server (URL to test http server stored in svc.SelfLink field)
func (h *testMetricsHandler) GetRawMetrics(apiClient kubernetes.Interface, namespace, functionName string) ([]byte, error) {
	svc, err := apiClient.CoreV1().Services(namespace).Get(functionName, metav1.GetOptions{})
	if err != nil {
		return []byte{}, err
	}
	b, err := http.Get(svc.SelfLink)
	if err != nil {
		return nil, err
	}
	defer b.Body.Close()
	return ioutil.ReadAll(b.Body)
}

func topOutput(t *testing.T, client versioned.Interface, apiV1Client kubernetes.Interface, h utils.MetricsRetriever, ns, functionName, output string) string {
	var buf bytes.Buffer

	if err := doTop(&buf, client, apiV1Client, h, ns, functionName, output); err != nil {
		t.Fatalf("doTop returned error: %v", err)
	}

	return buf.String()
}

func TestTop(t *testing.T) {

	// setup test server to serve the /metrics endpoint
	ts2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w,
			`# HELP go_gc_duration_seconds A summary of the GC invocation durations.
			# TYPE go_gc_duration_seconds summary
			go_gc_duration_seconds{quantile="0"} 1.6846e-05
			go_gc_duration_seconds{quantile="0.25"} 3.9124e-05
			go_gc_duration_seconds{quantile="0.5"} 0.000147183
			go_gc_duration_seconds{quantile="0.75"} 0.000958419
			go_gc_duration_seconds{quantile="1"} 0.00796035
			go_gc_duration_seconds_sum 2.50781303
			go_gc_duration_seconds_count 3424
			# HELP go_goroutines Number of goroutines that currently exist.
			# TYPE go_goroutines gauge
			go_goroutines 7
			# HELP go_info Information about the Go environment.
			# TYPE go_info gauge
			go_info{version="go1.10.2"} 1
			# HELP go_memstats_alloc_bytes Number of bytes allocated and still in use.
			# TYPE go_memstats_alloc_bytes gauge
			go_memstats_alloc_bytes 2.28336e+06
			# HELP go_memstats_alloc_bytes_total Total number of bytes allocated, even if freed.
			# TYPE go_memstats_alloc_bytes_total counter
			go_memstats_alloc_bytes_total 9.9682544e+09
			# HELP go_memstats_buck_hash_sys_bytes Number of bytes used by the profiling bucket hash table.
			# TYPE go_memstats_buck_hash_sys_bytes gauge
			go_memstats_buck_hash_sys_bytes 1.500081e+06
			# HELP go_memstats_frees_total Total number of frees.
			# TYPE go_memstats_frees_total counter
			go_memstats_frees_total 1.2698678e+07
			# HELP go_memstats_gc_cpu_fraction The fraction of this program's available CPU time used by the GC since the program started.
			# TYPE go_memstats_gc_cpu_fraction gauge
			go_memstats_gc_cpu_fraction 0.0001214506861340198
			# HELP go_memstats_gc_sys_bytes Number of bytes used for garbage collection system metadata.
			# TYPE go_memstats_gc_sys_bytes gauge
			go_memstats_gc_sys_bytes 405504
			# HELP go_memstats_heap_alloc_bytes Number of heap bytes allocated and still in use.
			# TYPE go_memstats_heap_alloc_bytes gauge
			go_memstats_heap_alloc_bytes 2.28336e+06
			# HELP go_memstats_heap_idle_bytes Number of heap bytes waiting to be used.
			# TYPE go_memstats_heap_idle_bytes gauge
			go_memstats_heap_idle_bytes 2.6624e+06
			# HELP go_memstats_heap_inuse_bytes Number of heap bytes that are in use.
			# TYPE go_memstats_heap_inuse_bytes gauge
			go_memstats_heap_inuse_bytes 3.072e+06
			# HELP go_memstats_heap_objects Number of allocated objects.
			# TYPE go_memstats_heap_objects gauge
			go_memstats_heap_objects 6280
			# HELP go_memstats_heap_released_bytes Number of heap bytes released to OS.
			# TYPE go_memstats_heap_released_bytes gauge
			go_memstats_heap_released_bytes 0
			# HELP go_memstats_heap_sys_bytes Number of heap bytes obtained from system.
			# TYPE go_memstats_heap_sys_bytes gauge
			go_memstats_heap_sys_bytes 5.7344e+06
			# HELP go_memstats_last_gc_time_seconds Number of seconds since 1970 of last garbage collection.
			# TYPE go_memstats_last_gc_time_seconds gauge
			go_memstats_last_gc_time_seconds 1.528573398809276e+09
			# HELP go_memstats_lookups_total Total number of pointer lookups.
			# TYPE go_memstats_lookups_total counter
			go_memstats_lookups_total 88701
			# HELP go_memstats_mallocs_total Total number of mallocs.
			# TYPE go_memstats_mallocs_total counter
			go_memstats_mallocs_total 1.2704958e+07
			# HELP go_memstats_mcache_inuse_bytes Number of bytes in use by mcache structures.
			# TYPE go_memstats_mcache_inuse_bytes gauge
			go_memstats_mcache_inuse_bytes 3472
			# HELP go_memstats_mcache_sys_bytes Number of bytes used for mcache structures obtained from system.
			# TYPE go_memstats_mcache_sys_bytes gauge
			go_memstats_mcache_sys_bytes 16384
			# HELP go_memstats_mspan_inuse_bytes Number of bytes in use by mspan structures.
			# TYPE go_memstats_mspan_inuse_bytes gauge
			go_memstats_mspan_inuse_bytes 25688
			# HELP go_memstats_mspan_sys_bytes Number of bytes used for mspan structures obtained from system.
			# TYPE go_memstats_mspan_sys_bytes gauge
			go_memstats_mspan_sys_bytes 32768
			# HELP go_memstats_next_gc_bytes Number of heap bytes when next garbage collection will take place.
			# TYPE go_memstats_next_gc_bytes gauge
			go_memstats_next_gc_bytes 4.194304e+06
			# HELP go_memstats_other_sys_bytes Number of bytes used for other system allocations.
			# TYPE go_memstats_other_sys_bytes gauge
			go_memstats_other_sys_bytes 738631
			# HELP go_memstats_stack_inuse_bytes Number of bytes in use by the stack allocator.
			# TYPE go_memstats_stack_inuse_bytes gauge
			go_memstats_stack_inuse_bytes 557056
			# HELP go_memstats_stack_sys_bytes Number of bytes obtained from system for stack allocator.
			# TYPE go_memstats_stack_sys_bytes gauge
			go_memstats_stack_sys_bytes 557056
			# HELP go_memstats_sys_bytes Number of bytes obtained from system.
			# TYPE go_memstats_sys_bytes gauge
			go_memstats_sys_bytes 8.984824e+06
			# HELP go_threads Number of OS threads created.
			# TYPE go_threads gauge
			go_threads 10
			# HELP process_cpu_seconds_total Total user and system CPU time spent in seconds.
			# TYPE process_cpu_seconds_total counter
			process_cpu_seconds_total 25.88
			# HELP process_max_fds Maximum number of open file descriptors.
			# TYPE process_max_fds gauge
			process_max_fds 1.048576e+06
			# HELP process_open_fds Number of open file descriptors.
			# TYPE process_open_fds gauge
			process_open_fds 8
			# HELP process_resident_memory_bytes Resident memory size in bytes.
			# TYPE process_resident_memory_bytes gauge
			process_resident_memory_bytes 1.3942784e+07
			# HELP process_start_time_seconds Start time of the process since unix epoch in seconds.
			# TYPE process_start_time_seconds gauge
			process_start_time_seconds 1.52853941225e+09
			# HELP process_virtual_memory_bytes Virtual memory size in bytes.
			# TYPE process_virtual_memory_bytes gauge
			process_virtual_memory_bytes 1.57294592e+08
			# HELP promhttp_metric_handler_requests_in_flight Current number of scrapes being served.
			# TYPE promhttp_metric_handler_requests_in_flight gauge
			promhttp_metric_handler_requests_in_flight 1
			# HELP promhttp_metric_handler_requests_total Total number of scrapes by HTTP status code.
			# TYPE promhttp_metric_handler_requests_total counter
			promhttp_metric_handler_requests_total{code="200"} 10798
			promhttp_metric_handler_requests_total{code="500"} 0
			promhttp_metric_handler_requests_total{code="503"} 0

`)
	}))
	defer ts2.Close()

	// setup test server to serve the /metrics endpoint
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w,
			`# HELP process_virtual_memory_bytes Virtual memory size in bytes.
			# TYPE process_virtual_memory_bytes gauge
			process_virtual_memory_bytes 815255552.0
			# HELP process_resident_memory_bytes Resident memory size in bytes.
			# TYPE process_resident_memory_bytes gauge
			process_resident_memory_bytes 25001984.0
			# HELP process_start_time_seconds Start time of the process since unix epoch in seconds.
			# TYPE process_start_time_seconds gauge
			process_start_time_seconds 1528507334.03
			# HELP process_cpu_seconds_total Total user and system CPU time spent in seconds.
			# TYPE process_cpu_seconds_total counter
			process_cpu_seconds_total 54.72
			# HELP process_open_fds Number of open file descriptors.
			# TYPE process_open_fds gauge
			process_open_fds 8.0
			# HELP process_max_fds Maximum number of open file descriptors.
			# TYPE process_max_fds gauge
			process_max_fds 1048576.0
			# HELP python_info Python platform information
			# TYPE python_info gauge
			python_info{implementation="CPython",major="2",minor="7",patchlevel="9",version="2.7.9"} 1.0
			# HELP function_failures_total Number of exceptions in user function
			# TYPE function_failures_total counter
			function_failures_total{method="GET"} 0.0
			function_failures_total{method="POST"} 0.0
			# HELP function_calls_total Number of calls to user function
			# TYPE function_calls_total counter
			function_calls_total{method="GET"} 254.0
			function_calls_total{method="POST"} 296.0
			# HELP function_duration_seconds Duration of user function in seconds
			# TYPE function_duration_seconds histogram
			function_duration_seconds_bucket{le="0.005",method="GET"} 8.0
			function_duration_seconds_bucket{le="0.01",method="GET"} 191.0
			function_duration_seconds_bucket{le="0.025",method="GET"} 248.0
			function_duration_seconds_bucket{le="0.05",method="GET"} 253.0
			function_duration_seconds_bucket{le="0.075",method="GET"} 253.0
			function_duration_seconds_bucket{le="0.1",method="GET"} 253.0
			function_duration_seconds_bucket{le="0.25",method="GET"} 254.0
			function_duration_seconds_bucket{le="0.5",method="GET"} 254.0
			function_duration_seconds_bucket{le="0.75",method="GET"} 254.0
			function_duration_seconds_bucket{le="1.0",method="GET"} 254.0
			function_duration_seconds_bucket{le="2.5",method="GET"} 254.0
			function_duration_seconds_bucket{le="5.0",method="GET"} 254.0
			function_duration_seconds_bucket{le="7.5",method="GET"} 254.0
			function_duration_seconds_bucket{le="10.0",method="GET"} 254.0
			function_duration_seconds_bucket{le="+Inf",method="GET"} 254.0
			function_duration_seconds_count{method="GET"} 254.0
			function_duration_seconds_sum{method="GET"} 2.863368272781372
			function_duration_seconds_bucket{le="0.005",method="POST"} 1.0
			function_duration_seconds_bucket{le="0.01",method="POST"} 157.0
			function_duration_seconds_bucket{le="0.025",method="POST"} 296.0
			function_duration_seconds_bucket{le="0.05",method="POST"} 296.0
			function_duration_seconds_bucket{le="0.075",method="POST"} 296.0
			function_duration_seconds_bucket{le="0.1",method="POST"} 296.0
			function_duration_seconds_bucket{le="0.25",method="POST"} 296.0
			function_duration_seconds_bucket{le="0.5",method="POST"} 296.0
			function_duration_seconds_bucket{le="0.75",method="POST"} 296.0
			function_duration_seconds_bucket{le="1.0",method="POST"} 296.0
			function_duration_seconds_bucket{le="2.5",method="POST"} 296.0
			function_duration_seconds_bucket{le="5.0",method="POST"} 296.0
			function_duration_seconds_bucket{le="7.5",method="POST"} 296.0
			function_duration_seconds_bucket{le="10.0",method="POST"} 296.0
			function_duration_seconds_bucket{le="+Inf",method="POST"} 296.0
			function_duration_seconds_count{method="POST"} 296.0
			function_duration_seconds_sum{method="POST"} 3.4116291999816895

`)
	}))
	defer ts.Close()

	function1Name := "pyFunc"
	function2Name := "goFunc"
	namespace := "myns"

	listObj := kubelessApi.FunctionList{
		Items: []*kubelessApi.Function{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      function1Name,
					Namespace: namespace,
				},
				Spec: kubelessApi.FunctionSpec{
					Handler:  "fhandler",
					Function: function1Name,
					Runtime:  "pyruntime",
					Deps:     "pydeps",
					Deployment: v1beta1.Deployment{
						Spec: v1beta1.DeploymentSpec{
							Template: v1.PodTemplateSpec{
								Spec: v1.PodSpec{
									Containers: []v1.Container{{}},
								},
							},
						},
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      function2Name,
					Namespace: namespace,
				},
				Spec: kubelessApi.FunctionSpec{
					Handler:  "gohandler",
					Function: function2Name,
					Runtime:  "goruntime",
					Deps:     "godeps",
					Deployment: v1beta1.Deployment{
						Spec: v1beta1.DeploymentSpec{
							Template: v1.PodTemplateSpec{
								Spec: v1.PodSpec{
									Containers: []v1.Container{{}},
								},
							},
						},
					},
				},
			},
		},
	}

	client := fFake.NewSimpleClientset(listObj.Items[0], listObj.Items[1])

	deploymentPy := v1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      function1Name,
			Namespace: namespace,
		},
		Status: v1beta1.DeploymentStatus{
			Replicas:      int32(1),
			ReadyReplicas: int32(1),
		},
	}
	deploymentGo := v1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      function2Name,
			Namespace: namespace,
		},
		Status: v1beta1.DeploymentStatus{
			Replicas:      int32(1),
			ReadyReplicas: int32(1),
		},
	}
	serviceGo := v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      function2Name,
			Namespace: namespace,
			SelfLink:  ts2.URL,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:       "p1",
					Port:       int32(8080),
					TargetPort: intstr.FromInt(8080),
					NodePort:   0,
					Protocol:   v1.ProtocolTCP,
				},
			},
		},
	}
	servicePy := v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      function1Name,
			Namespace: namespace,
			SelfLink:  ts.URL,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:       "p1",
					Port:       int32(8080),
					TargetPort: intstr.FromInt(8080),
					NodePort:   0,
					Protocol:   v1.ProtocolTCP,
				},
			},
		},
	}

	apiV1Client := fake.NewSimpleClientset(&deploymentPy, &servicePy, &deploymentGo, &serviceGo)

	handler := &testMetricsHandler{}

	// List multiple functions
	output := topOutput(t, client, apiV1Client, handler, namespace, "", "")
	t.Log("output is", output)

	if !strings.Contains(output, function1Name) || !strings.Contains(output, function2Name) || !strings.Contains(output, namespace) {
		t.Errorf("table output didn't match FUNCTION or NAMESPACE")
	}
	if !strings.Contains(output, "GET") || !strings.Contains(output, "POST") {
		t.Errorf("table output didn't match on METHOD")
	}
	if !strings.Contains(output, "2.86336") || !strings.Contains(output, "3.41162") {
		t.Errorf("table output didn't match on TOTAL_DURATION_SECONDS")
	}
	// verify calculated fields
	if !strings.Contains(output, "0.0112731") || !strings.Contains(output, "0.0115257") {
		t.Errorf("table output didn't match on AVG_DURATION_SECONDS")
	}

	// Get single function
	output = topOutput(t, client, apiV1Client, handler, namespace, function2Name, "")
	t.Log("output is", output)

	if strings.Contains(output, function1Name) || !strings.Contains(output, function2Name) || !strings.Contains(output, namespace) {
		t.Errorf("table output didn't match FUNCTION or NAMESPACE")
	}

	// json output
	output = topOutput(t, client, apiV1Client, handler, namespace, "", "json")
	t.Log("output is", output)

	if !strings.Contains(output, function1Name) || !strings.Contains(output, function2Name) || !strings.Contains(output, namespace) {
		t.Errorf("table output didn't match FUNCTION or NAMESPACE")
	}
	if !strings.Contains(output, "GET") || !strings.Contains(output, "POST") {
		t.Errorf("table output didn't match on METHOD")
	}
	if !strings.Contains(output, "2.86336") || !strings.Contains(output, "3.41162") {
		t.Errorf("table output didn't match on TOTAL_DURATION_SECONDS")
	}
	// verify calculated fields
	if !strings.Contains(output, "0.0112731") || !strings.Contains(output, "0.0115257") {
		t.Errorf("table output didn't match on AVG_DURATION_SECONDS")
	}

	// yaml output
	output = topOutput(t, client, apiV1Client, handler, namespace, "", "yaml")
	t.Log("output is", output)

	if !strings.Contains(output, function1Name) || !strings.Contains(output, function2Name) || !strings.Contains(output, namespace) {
		t.Errorf("table output didn't match FUNCTION or NAMESPACE")
	}
	if !strings.Contains(output, "GET") || !strings.Contains(output, "POST") {
		t.Errorf("table output didn't match on METHOD")
	}
	if !strings.Contains(output, "2.86336") || !strings.Contains(output, "3.41162") {
		t.Errorf("table output didn't match on TOTAL_DURATION_SECONDS")
	}
	// verify calculated fields
	if !strings.Contains(output, "0.0112731") || !strings.Contains(output, "0.0115257") {
		t.Errorf("table output didn't match on AVG_DURATION_SECONDS")
	}

}
