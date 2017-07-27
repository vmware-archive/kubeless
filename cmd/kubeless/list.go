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
	"io"

	"github.com/Sirupsen/logrus"
	"github.com/ghodss/yaml"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/rest"

	"github.com/kubeless/kubeless/pkg/spec"
	"github.com/kubeless/kubeless/pkg/utils"
)

var listCmd = &cobra.Command{
	Use:     "list FLAG",
	Aliases: []string{"ls"},
	Short:   "list all functions deployed to Kubeless",
	Long:    `list all functions deployed to Kubeless`,
	Run: func(cmd *cobra.Command, args []string) {
		output, err := cmd.Flags().GetString("out")
		if err != nil {
			logrus.Fatal(err.Error())
		}
		ns, err := cmd.Flags().GetString("namespace")
		if err != nil {
			logrus.Fatal(err.Error())
		}

		tprClient, err := utils.GetTPRClientOutOfCluster()
		if err != nil {
			logrus.Fatalf("Can not list functions: %v", err)
		}

		if err := doList(cmd.OutOrStdout(), tprClient, ns, output, args); err != nil {
			logrus.Fatal(err.Error())
		}
	},
}

func init() {
	listCmd.Flags().StringP("out", "o", "", "Output format. One of: json|yaml")
	listCmd.Flags().StringP("namespace", "n", api.NamespaceDefault, "Specify namespace for the function")
}

func doList(w io.Writer, tprClient rest.Interface, ns, output string, args []string) error {
	var list []*spec.Function
	if len(args) == 0 {
		funcList := spec.FunctionList{}
		err := tprClient.Get().
			Resource("functions").
			Namespace(ns).
			Do().
			Into(&funcList)
		if err != nil {
			return err
		}
		list = funcList.Items
	} else {
		list = make([]*spec.Function, 0, len(args))
		for _, arg := range args {
			f := spec.Function{}
			err := tprClient.Get().
				Resource("functions").
				Namespace(ns).
				Name(arg).
				Do().
				Into(&f)
			if err != nil {
				return fmt.Errorf("Error listing function %s: %v", arg, err)
			}
			list = append(list, &f)
		}
	}

	return printFunctions(w, list, output)
}

// printFunctions formats the output of function list
func printFunctions(w io.Writer, functions []*spec.Function, output string) error {
	if output == "" {
		table := tablewriter.NewWriter(w)
		table.SetHeader([]string{"Name", "description", "namespace", "handler", "runtime", "type", "topic", "dependencies", "memory", "env"})
		for _, f := range functions {
			n := f.Metadata.Name
			desc := f.Spec.Desc
			h := f.Spec.Handler
			r := f.Spec.Runtime
			t := f.Spec.Type
			tp := f.Spec.Topic
			ns := f.Metadata.Namespace
			dep := f.Spec.Deps
			mem := f.Spec.Template.Spec.Containers[0].Resources.Requests.Memory().String()
			env, _ := json.Marshal(f.Spec.Template.Spec.Containers[0].Env)
			table.Append([]string{n, desc, ns, h, r, t, tp, dep, mem, string(env)})
		}
		table.Render()
	} else {
		for _, f := range functions {
			switch output {
			case "json":
				b, _ := json.MarshalIndent(f, "", "  ")
				fmt.Fprintln(w, string(b))
			case "yaml":
				b, _ := yaml.Marshal(f)
				fmt.Fprintln(w, string(b))
			default:
				return fmt.Errorf("Wrong output format. Please use only json|yaml")
			}
		}
	}
	return nil
}
