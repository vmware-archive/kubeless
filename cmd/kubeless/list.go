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
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/bitnami/kubeless/pkg/controller"
	"github.com/bitnami/kubeless/pkg/spec"
	"github.com/bitnami/kubeless/pkg/utils"
)

var listCmd = &cobra.Command{
	Use:     "list FLAG",
	Aliases: []string{"ls"},
	Short:   "list all functions deployed to Kubeless",
	Long:    `list all functions deployed to Kubeless`,
	Run: func(cmd *cobra.Command, args []string) {
		output, err := cmd.Flags().GetString("out")
		cfg := controller.Config{
			KubeCli: utils.GetClientOutOfCluster(),
		}
		c := controller.New(cfg)

		tprClient, err := utils.GetTPRClientOutOfCluster()
		if err != nil {
			logrus.Fatalf("Can not list functions: %v", err)
		}
		_, err = c.FindResourceVersion(tprClient)
		if err != nil {
			logrus.Fatalf("Can not list functions: %v", err)
		}

		if len(args) == 0 {
			for k := range c.Functions {
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

// printFunctions formats the output of function list
func printFunctions(args []string, functions map[string]*spec.Function, output string) {
	if output == "" {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Name", "namespace", "handler", "runtime", "type", "topic", "dependencies"})
		for _, f := range args {
			n := strings.Split(fmt.Sprintf(f), ".")[0]
			h := fmt.Sprintf(functions[f].Spec.Handler)
			r := fmt.Sprintf(functions[f].Spec.Runtime)
			t := fmt.Sprintf(functions[f].Spec.Type)
			tp := fmt.Sprintf(functions[f].Spec.Topic)
			ns := fmt.Sprintf(functions[f].Metadata.Namespace)
			dep := fmt.Sprintf(functions[f].Spec.Deps)
			table.Append([]string{n, ns, h, r, t, tp, dep})
		}
		table.Render()
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
				fmt.Println("Wrong output format. Please use only json|yaml")
			}
		}
	}
}
