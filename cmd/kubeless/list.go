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
	"encoding/json"
	"fmt"

	"github.com/gosuri/uitable"
	"github.com/skippbox/kubeless/pkg/controller"
	"github.com/skippbox/kubeless/pkg/spec"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var listCmd = &cobra.Command{
	Use:   "ls FLAG",
	Short: "list all functions deployed to Kubeless",
	Long:  `list all functions deployed to Kubeless`,
	Run: func(cmd *cobra.Command, args []string) {
		output, err := cmd.Flags().GetString("out")
		//ns, err := cmd.Flags().GetString("namespace")
		cfg := controller.NewControllerConfig("", "")
		c := controller.New(cfg)
		_, err = c.FindResourceVersion()
		if err != nil {
			fmt.Errorf("Can not list functions: %v", err)
		}

		if len(args) == 0 {
			for k, _ := range c.Functions {
				args = append(args, k)
			}
		}

		printFunctions(args, c.Functions, output)
	},
}

func init() {
	listCmd.Flags().StringP("out", "o", "", "Output format. One of: json|yaml")
	// TODO: list all namespaces now. Will add specific ns later
	//listCmd.Flags().StringP("namespace", "", "", "Specify namespace for the function")
}

func printFunctions(args []string, functions map[string]*spec.Function, output string) {
	if output == "" {
		table := uitable.New()
		table.MaxColWidth = 30
		table.AddRow("NAME", "HANDLER", "RUNTIME", "TYPE", "TOPIC", "NAMESPACE")

		for _, f := range args {
			n := fmt.Sprintf(f)
			h := fmt.Sprintf(functions[f].Spec.Handler)
			r := fmt.Sprintf(functions[f].Spec.Runtime)
			t := fmt.Sprintf(functions[f].Spec.Type)
			tp := fmt.Sprintf(functions[f].Spec.Topic)
			ns := fmt.Sprintf(functions[f].ObjectMeta.Namespace)
			table.AddRow(n, h, r, t, tp, ns)
		}
		fmt.Println(table.String())
	} else {
		for _, f := range args {
			switch output {
			case "json":
				b, _ := json.MarshalIndent(functions[f].Spec, "", "  ")
				fmt.Println(string(b))
			case "yaml":
				b, _ := yaml.Marshal(functions[f].Spec)
				fmt.Println(string(b))
			default:
				fmt.Errorf("Wrong output format. Please use only json|yaml.")
			}
		}
	}
}
