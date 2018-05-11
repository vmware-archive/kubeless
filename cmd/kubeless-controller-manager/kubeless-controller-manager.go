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

// Kubeless controller binary.
//
// See github.com/kubeless/kubeless/pkg/controller
package main

import (
	"fmt"
	"os"
	"net"
	"os/signal"
	"syscall"
	"strconv"
	"runtime"
	"net/http"
	"net/http/pprof"

	"github.com/prometheus/client_golang/prometheus"
	monitoringv1alpha1 "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1alpha1"
	"github.com/kubeless/kubeless/pkg/client/informers/externalversions"
	"github.com/kubeless/kubeless/pkg/controller"
	"github.com/kubeless/kubeless/pkg/utils"
	"github.com/kubeless/kubeless/pkg/version"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
)

const (
	globalUsage = `` //TODO: adding explanation
)

var rootCmd = &cobra.Command{
	Use:   "kubeless-controller",
	Short: "Kubeless controller",
	Long:  globalUsage,
	Run: func(cmd *cobra.Command, args []string) {
		kubelessClient, err := utils.GetFunctionClientInCluster()
		if err != nil {
			logrus.Fatalf("Cannot get kubeless client: %v", err)
		}

		functionCfg := controller.Config{
			KubeCli:        utils.GetClient(),
			FunctionClient: kubelessClient,
		}
		httpTriggerCfg := controller.HTTPTriggerConfig{
			KubeCli:       utils.GetClient(),
			TriggerClient: kubelessClient,
		}
		cronJobTriggerCfg := controller.CronJobTriggerConfig{
			KubeCli:       utils.GetClient(),
			TriggerClient: kubelessClient,
		}

		restCfg, err := rest.InClusterConfig()
		if err != nil {
			logrus.Fatalf("Cannot get REST client: %v", err)
		}
		// ServiceMonitor client is needed for handling monitoring resources
		smclient, err := monitoringv1alpha1.NewForConfig(restCfg)
		if err != nil {
			logrus.Fatal(err)
		}

		sharedInformerFactory := externalversions.NewSharedInformerFactory(kubelessClient, 0)

		functionController := controller.NewFunctionController(functionCfg, smclient)
		httpTriggerController := controller.NewHTTPTriggerController(httpTriggerCfg, sharedInformerFactory)
		cronJobTriggerController := controller.NewCronJobTriggerController(cronJobTriggerCfg, sharedInformerFactory)

		port, err := cmd.Flags().GetUint16("port")
		if err != nil {
			logrus.Fatal("Invalid Server Port")
		}

		address, err := cmd.Flags().GetString("addresss")
		if err != nil {
			logrus.Fatal("Cannot get Server Address")
		}

		isEnableProfiling, err := cmd.Flags().GetBool("profiling")
		if err != nil {
			logrus.Fatal("Cannot get profiling")
		}

		isEnableContentionProfiling, err := cmd.Flags().GetBool("contention-profiling")
		if err != nil {
			logrus.Fatal("Cannot get contention-profiling")
		}

		go func() {
			mux := http.NewServeMux()
			if isEnableProfiling {
				mux.HandleFunc("/debug/pprof/", pprof.Index)
				mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
				mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
				mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
				if isEnableContentionProfiling {
					runtime.SetBlockProfileRate(1)
				}
			}
			mux.Handle("/metrics", prometheus.Handler())

			server := &http.Server{
				Addr:    net.JoinHostPort(address, strconv.Itoa(int(port))),
				Handler: mux,
			}
			logrus.Fatal(server.ListenAndServe())
		}()

		stopCh := make(chan struct{})
		defer close(stopCh)

		go functionController.Run(stopCh)
		go httpTriggerController.Run(stopCh)
		go cronJobTriggerController.Run(stopCh)

		sigterm := make(chan os.Signal, 1)
		signal.Notify(sigterm, syscall.SIGTERM)
		signal.Notify(sigterm, syscall.SIGINT)
		<-sigterm
	},
}

func init() {
	rootCmd.Flags().StringP("address", "", "0.0.0.0", "The IP address to serve on (set to 0.0.0.0 for all interfaces)")
	rootCmd.Flags().Uint16P("port", "", 10301, "The port that the controller-manager's http service runs on")
	rootCmd.Flags().BoolP("profiling", "", false, "Enable profiling via web interface host:port/debug/pprof/")
	rootCmd.Flags().BoolP("contention-profiling", "", false, "Enable lock contention profiling, if profiling is enabled")
}

func main() {
	logrus.Infof("Running Kubeless controller manager version: %v", version.Version)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
