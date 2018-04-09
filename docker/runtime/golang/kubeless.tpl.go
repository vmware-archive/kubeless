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

package main

import (
	"fmt"
	"golang.org/x/net/context"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"kubeless"

	"github.com/kubeless/kubeless/pkg/functions"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	funcHandler   = os.Getenv("FUNC_HANDLER")
	timeout       = os.Getenv("FUNC_TIMEOUT")
	funcPort      = os.Getenv("FUNC_PORT")
	runtime       = os.Getenv("FUNC_RUNTIME")
	memoryLimit   = os.Getenv("FUNC_MEMORY_LIMIT")
	intTimeout    int
	funcContext   functions.Context
	funcHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "function_duration_seconds",
		Help: "Duration of user function in seconds",
	}, []string{"method"})
	funcCalls = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "function_calls_total",
		Help: "Number of calls to user function",
	}, []string{"method"})
	funcErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "function_failures_total",
		Help: "Number of exceptions in user function",
	}, []string{"method"})
)

func init() {
	if timeout == "" {
		timeout = "180"
	}
	if funcPort == "" {
		funcPort = "8080"
	}
	funcContext = functions.Context{
		FunctionName: funcHandler,
		Timeout:      timeout,
		Runtime:      runtime,
		MemoryLimit:  memoryLimit,
	}
	var err error
	intTimeout, err = strconv.Atoi(timeout)
	if err != nil {
		panic(err)
	}
	prometheus.MustRegister(funcHistogram, funcCalls, funcErrors)
}

// Logging Functions, required to expose statusCode property
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func logReq(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lrw := newLoggingResponseWriter(w)
		handler.ServeHTTP(lrw, r)
		log.Printf("%s \"%s %s %s\" %d %s", r.RemoteAddr, r.Method, r.RequestURI, r.Proto, lrw.statusCode, r.UserAgent())
		if lrw.statusCode == 408 {
			go func() {
				// Give time to return timeout response
				time.Sleep(time.Second)
				log.Fatal("Request timeout. Forcing exit")
			}()
		}
	})
}

// Handling Functions
func health(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func handler(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(intTimeout)*time.Second)
	defer cancel()
	event := functions.Event{
		Data:           string(data),
		EventID:        r.Header.Get("event-id"),
		EventType:      r.Header.Get("event-type"),
		EventTime:      r.Header.Get("event-time"),
		EventNamespace: r.Header.Get("event-namespace"),
		Extensions: functions.Extension{
			Request: r,
			Context: ctx,
		},
	}
	funcChannel := make(chan struct {
		res string
		err error
	}, 1)
	go func() {
		funcCalls.With(prometheus.Labels{"method": r.Method}).Inc()
		start := time.Now()
		res, err := kubeless.<<FUNCTION>>(event, funcContext)
		funcHistogram.With(prometheus.Labels{"method": r.Method}).Observe(time.Since(start).Seconds())
		pack := struct {
			res string
			err error
		}{res, err}
		funcChannel <- pack
	}()
	select {
	case respPack := <-funcChannel:
		if respPack.err != nil {
			funcErrors.With(prometheus.Labels{"method": r.Method}).Inc()
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Error: %v", respPack.err)))
		} else {
			w.Write([]byte(respPack.res))
		}
	// Send Timeout response
	case <-ctx.Done():
		funcErrors.With(prometheus.Labels{"method": r.Method}).Inc()
		w.WriteHeader(http.StatusRequestTimeout)
		w.Write([]byte("Timeout exceeded"))
	}
}

func main() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/healthz", health)
	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(fmt.Sprintf(":%s", funcPort), logReq(http.DefaultServeMux)); err != nil {
		panic(err)
	}
}
