/*
Copyright 2016 Skippbox, Ltd.

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

package cmd

import (
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/skippbox/kubeless/pkg/controller"
	"github.com/skippbox/kubeless/pkg/utils"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "kubeless",
	Short: "Serverless framework for Kubernetes",
	Long:  `Serverless framework for Kubernetes`,
	Run: func(cmd *cobra.Command, args []string) {
		master, err := cmd.Flags().GetString("master")
		if master == "" {
			master = "localhost"
		}
		cfg := newControllerConfig(master, "")
		c := controller.New(cfg)
		err = c.Run()
		if err != nil {
			logrus.Fatalf("Kubeless controller running failed: %s", err)
		}
	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	RootCmd.Flags().StringP("master", "", "", "Apiserver address")
}

func newControllerConfig(masterHost, ns string) controller.Config {
	f := utils.GetFactory()
	kubecli, err := f.Client()
	if err != nil {
		fmt.Errorf("Can not get kubernetes config: %s", err)
	}
	if ns == "" {
		ns, _, err = f.DefaultNamespace()
		if err != nil {
			fmt.Errorf("Can not get kubernetes config: %s", err)
		}
	}
	if masterHost == "" {
		k8sConfig, err := f.ClientConfig()
		if err != nil {
			fmt.Errorf("Can not get kubernetes config: %s", err)
		}
		masterHost = k8sConfig.Host
	}
	cfg := controller.Config{
		Namespace:  ns,
		KubeCli:    kubecli,
		MasterHost: masterHost,
	}

	return cfg
}
