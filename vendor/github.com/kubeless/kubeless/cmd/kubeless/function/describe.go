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

package function

import (
	"encoding/json"
	"fmt"

	"github.com/ghodss/yaml"
	"github.com/gosuri/uitable"
	kubelessApi "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	"github.com/kubeless/kubeless/pkg/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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
		if ns == "" {
			ns = utils.GetDefaultNamespace()
		}

		output, err := cmd.Flags().GetString("out")
		if err != nil {
			logrus.Fatalf("Can not describe function: %v", err)
		}

		f, err := utils.GetFunction(funcName, ns)
		if err != nil {
			logrus.Fatalf("Can not describe function: %v", err)
		}

		err = print(f, funcName, output)
		if err != nil {
			logrus.Fatalf("Can not describe function: %v", err)
		}
	},
}

func init() {
	describeCmd.Flags().StringP("out", "o", "", "Output format. One of: json|yaml")
	describeCmd.Flags().StringP("namespace", "n", "", "Specify namespace for the function")
}

func print(f kubelessApi.Function, name, output string) error {
	switch output {
	case "":
		table := uitable.New()
		table.MaxColWidth = 80
		table.Wrap = true
		label, err := json.Marshal(f.ObjectMeta.Labels)
		if err != nil {
			return err
		}
		var env, memory string
		if len(f.Spec.Deployment.Spec.Template.Spec.Containers) != 0 {
			b, err := json.Marshal(f.Spec.Deployment.Spec.Template.Spec.Containers[0].Env)
			if err != nil {
				return err
			}
			env = string(b)
			memory = f.Spec.Deployment.Spec.Template.Spec.Containers[0].Resources.Requests.Memory().String()
		}

		table.AddRow("Name:", name)
		table.AddRow("Namespace:", f.ObjectMeta.Namespace)
		table.AddRow("Handler:", f.Spec.Handler)
		table.AddRow("Runtime:", f.Spec.Runtime)
		table.AddRow("Label:", string(label))
		table.AddRow("Envvar:", env)
		table.AddRow("Memory:", memory)
		table.AddRow("Dependencies:", f.Spec.Deps)
		fmt.Println(table)
	case "json":
		b, err := json.MarshalIndent(f, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(b))
	case "yaml":
		b, err := yaml.Marshal(f)
		if err != nil {
			return err
		}
		fmt.Println(string(b))
	default:
		fmt.Println("Wrong output format. Please use only json|yaml")
	}

	return nil
}
