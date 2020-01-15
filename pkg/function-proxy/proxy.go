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
	"golang.org/x/net/context"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/kubeless/kubeless/pkg/function-proxy/utils"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func copyHeaders(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func handle(ctx context.Context, w http.ResponseWriter, r *http.Request) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest(r.Method, "http://localhost:8090", r.Body)
	if err != nil {
		return []byte{}, err
	}
	copyHeaders(req.Header, r.Header)
	req.ContentLength = r.ContentLength
	response, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	}
	return ioutil.ReadAll(response.Body)
}

func handler(w http.ResponseWriter, r *http.Request) {
	utils.Handler(w, r, handle)
}

func health(w http.ResponseWriter, r *http.Request) {
	rr, err := http.Get("http://localhost:8090/healthz")
	res, _ := ioutil.ReadAll(rr.Body)
	log.Println(string(res))
	if err != nil {
		log.Fatalln("localhost:8090 not responding")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server error"))
	} else {
		w.Write([]byte("OK"))
	}
}

func startNativeDaemon() {
	args := os.Getenv("FUNC_PROCESS")
	cmd := exec.Command("/bin/sh", "-c", args)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Fatalf("Unable to run %s. Received %v", args, err)
	}
}

func gracefulShutdown(server *http.Server) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	timeout := 10 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	log.Printf("Shuting down with timeout: %s\n", timeout)
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		log.Println("Server stopped")
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)
	mux.HandleFunc("/healthz", health)
	mux.Handle("/metrics", promhttp.Handler())

	server := utils.NewServer(mux)

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			panic(err)
		}
	}()

	gracefulShutdown(server)
}
