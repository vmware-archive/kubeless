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
	"fmt"
	"golang.org/x/net/context"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	timeout       = os.Getenv("FUNC_TIMEOUT")
	funcPort      = os.Getenv("FUNC_PORT")
	intTimeout    int
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

func copyHeaders(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

// Handle type receive the context elements of a HTTP request to process it
type Handle func(ctx context.Context, w http.ResponseWriter, r *http.Request) ([]byte, error)

// Handler receives an HTTP request and response and a handler function
// It manages timeouts and prometheus metrics
func Handler(w http.ResponseWriter, r *http.Request, h Handle) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(intTimeout)*time.Second)
	defer cancel()
	funcChannel := make(chan struct {
		res string
		err error
	}, 1)
	go func() {
		funcCalls.With(prometheus.Labels{"method": r.Method}).Inc()
		start := time.Now()
		res, err := h(ctx, w, r)
		funcHistogram.With(prometheus.Labels{"method": r.Method}).Observe(time.Since(start).Seconds())
		pack := struct {
			res string
			err error
		}{string(res), err}
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

// ListenAndServe starts an HTTP server in FUNC_PORT using custom logging
func ListenAndServe() {
	if err := http.ListenAndServe(fmt.Sprintf(":%s", funcPort), logReq(http.DefaultServeMux)); err != nil {
		panic(err)
	}
}
