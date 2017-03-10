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

package main

import (
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/skippbox/kubeless/pkg/controller"
	"github.com/spf13/cobra"
)

const globalUsage = `` //TODO: adding explanation

var rootCmd = &cobra.Command{
	Use:   "kubeless-controller",
	Short: "Kubeless controller",
	Long:  globalUsage,
	Run: func(cmd *cobra.Command, args []string) {
		master, err := cmd.Flags().GetString("master")
		if master == "" {
			master = "localhost"
		}
		cfg := controller.NewControllerConfig(master, "")
		c := controller.New(cfg)
		err = c.Run()
		if err != nil {
			logrus.Fatalf("Kubeless controller running failed: %s", err)
		}
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringP("master", "", "", "Apiserver address")
}
