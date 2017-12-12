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
	"os/signal"
	"syscall"

	monitoringv1alpha1 "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1alpha1"
	"github.com/kubeless/kubeless/pkg/controller"
	"github.com/kubeless/kubeless/pkg/utils"
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
		crdClient, err := utils.GetCRDClient()
		if err != nil {
			logrus.Fatalf("Cannot get CRD client: %v", err)
		}
		cfg := controller.Config{
			KubeCli:   utils.GetClient(),
			CRDClient: crdClient,
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

		c := controller.New(cfg, smclient)
		stopCh := make(chan struct{})
		defer close(stopCh)

		go c.Run(stopCh)

		sigterm := make(chan os.Signal, 1)
		signal.Notify(sigterm, syscall.SIGTERM)
		signal.Notify(sigterm, syscall.SIGINT)
		<-sigterm
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
