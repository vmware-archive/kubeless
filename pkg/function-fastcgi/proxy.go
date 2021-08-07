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

	"github.com/kubeless/kubeless/pkg/function-proxy/utils"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/yookoala/gofast"
)

func health(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
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

func main() {
	go startNativeDaemon()

	address := os.Getenv("FASTCGI_ADDR")
	connFactory := gofast.SimpleConnFactory("tcp", address)

	mux := http.NewServeMux()
	http.Handle("/", gofast.NewHandler(
		gofast.NewFileEndpoint(os.Getenv("FASTCGI_FILE"))(gofast.BasicSession),
		gofast.SimpleClientFactory(connFactory),
	))
	mux.HandleFunc("/healthz", health)
	mux.Handle("/metrics", promhttp.Handler())

	server := utils.NewServer(mux)

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			panic(err)
		}
	}()

	utils.GracefulShutdown(server)
}
