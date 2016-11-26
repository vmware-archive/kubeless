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
	"encoding/json"
	"fmt"

	"github.com/skippbox/kubeless/pkg/controller"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var listCmd = &cobra.Command{
	Use:   "ls FLAG",
	Short: "list all functions deployed to Kubeless",
	Long:  `list all functions deployed to Kubeless`,
	Run: func(cmd *cobra.Command, args []string) {
		master, err := cmd.Flags().GetString("master")
		if master == "" {
			master = "localhost"
		}
		output, err := cmd.Flags().GetString("out")
		cfg := newControllerConfig(master)
		c := controller.New(cfg)
		_, err = c.FindResourceVersion()
		if err != nil {
			fmt.Errorf("Can not list functions: %v", err)
		}
		for k, v := range c.Functions {
			switch output {
			case "":
				fmt.Println(k)
			case "json":
				b, _ := json.MarshalIndent(v.Spec, "", "  ")
				fmt.Println(string(b))
			case "yaml":
				b, _ := yaml.Marshal(v.Spec)
				fmt.Println(string(b))
			default:
				fmt.Errorf("Wrong output format. Please use only json|yaml.")
			}
		}
	},
}

func init() {
	listCmd.Flags().StringP("out", "o", "", "Output format. One of: json|yaml")
}
