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
	"os"
	"os/signal"
	"syscall"

	"github.com/kubeless/kubeless/pkg/controller"
	"github.com/kubeless/kubeless/pkg/utils"
	"github.com/kubeless/kubeless/pkg/version"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kinesis-controller",
	Short: "Kinesis controller",
	Long: `AWS Kinesis trigger CRD controller that watches for the creation/deletion/update events
				  of kinesisstrigger API object from the Kubernetes API server and creates/deletes Kinesis stream poller to
				  read records from the given stream. On reading a record from the stream fowards the record data to appropraiate
				  functions`,
	Run: func(cmd *cobra.Command, args []string) {
		kubelessClient, err := utils.GetFunctionClientInCluster()
		if err != nil {
			logrus.Fatalf("Cannot get kubeless client: %v", err)
		}

		kinesisTriggerCfg := controller.KinesisTriggerConfig{
			TriggerClient: kubelessClient,
		}

		kinesisTriggerController := controller.NewKinesisTriggerController(kinesisTriggerCfg)

		stopCh := make(chan struct{})
		defer close(stopCh)

		go kinesisTriggerController.Run(stopCh)

		sigterm := make(chan os.Signal, 1)
		signal.Notify(sigterm, syscall.SIGTERM)
		signal.Notify(sigterm, syscall.SIGINT)
		<-sigterm
	},
}

func main() {
	logrus.Infof("Running AWS Kinesis controller version: %v", version.Version)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
