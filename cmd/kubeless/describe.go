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
	"gopkg.in/yaml.v2"

	"github.com/Sirupsen/logrus"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/bitnami/kubeless/pkg/spec"
	"github.com/bitnami/kubeless/pkg/utils"
	"k8s.io/client-go/pkg/api"
	"os"
)

var describeCmd = &cobra.Command{
	Use:     "describe FLAG",
	Aliases: []string{"ls"},
	Short:   "describe a function deployed to Kubeless",
	Long:    `describe a function deployed to Kubeless`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			logrus.Fatal("Need exactly one argument - function name")
		}
		funcName := args[0]

		ns, err := cmd.Flags().GetString("namespace")
		if err != nil {
			logrus.Fatalf("Can not describe function: %v", err)
		}

		output, err := cmd.Flags().GetString("out")
		if err != nil {
			logrus.Fatalf("Can not describe function: %v", err)
		}

		f, err := utils.GetFunction(funcName, ns)
		if err != nil {
			logrus.Fatalf("Can not describe function: %v", err)
		}

		print(f, funcName, output)
	},
}

func init() {
	describeCmd.Flags().StringP("out", "o", "", "Output format. One of: json|yaml")
	describeCmd.Flags().StringP("namespace", "", api.NamespaceDefault, "Specify namespace for the function")
}

func print(f spec.Function, name, output string) {
	if output == "" {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Properties", "Value"})
		data := [][]string{
			{"Name", name},
			{"Namespace", fmt.Sprintf(f.Metadata.Namespace)},
			{"Handler", fmt.Sprintf(f.Spec.Handler)},
			{"Runtime", fmt.Sprintf(f.Spec.Runtime)},
			{"Type", fmt.Sprintf(f.Spec.Type)},
			{"Topic", fmt.Sprintf(f.Spec.Topic)},
			{"Dependencies", fmt.Sprintf(f.Spec.Deps)},
		}

		for _, v := range data {
			table.Append(v)
		}
		table.Render() // Send output
	} else {
		switch output {
		case "json":
			b, _ := json.MarshalIndent(f.Spec, "", "  ")
			fmt.Println(string(b))
		case "yaml":
			b, _ := yaml.Marshal(f.Spec)
			fmt.Println(string(b))
		default:
			fmt.Println("Wrong output format. Please use only json|yaml")
		}
	}
}
