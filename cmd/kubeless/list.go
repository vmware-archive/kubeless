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
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/Sirupsen/logrus"
	"github.com/ghodss/yaml"
	"github.com/gosuri/uitable"
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
		table := uitable.New()
		table.MaxColWidth = 50
		table.Wrap = true
		table.AddRow("NAME", "NAMESPACE", "HANDLER", "RUNTIME", "TYPE", "TOPIC")
		for _, f := range functions {
			n := f.Metadata.Name
			h := f.Spec.Handler
			r := f.Spec.Runtime
			t := f.Spec.Type
			tp := f.Spec.Topic
			ns := f.Metadata.Namespace
			table.AddRow(n, ns, h, r, t, tp)
		}
		fmt.Fprintln(w, table)
	} else if output == "wide" {
		table := uitable.New()
		table.MaxColWidth = 50
		table.Wrap = true
		table.AddRow("NAME", "NAMESPACE", "HANDLER", "RUNTIME", "TYPE", "TOPIC", "MEMORY", "ENV", "LABEL")
		for _, f := range functions {
			n := f.Metadata.Name
			h := f.Spec.Handler
			r := f.Spec.Runtime
			t := f.Spec.Type
			tp := f.Spec.Topic
			ns := f.Metadata.Namespace
			mem := ""
			env := ""
			if len(f.Spec.Template.Spec.Containers[0].Resources.Requests) != 0 {
				mem = f.Spec.Template.Spec.Containers[0].Resources.Requests.Memory().String()
			}
			if len(f.Spec.Template.Spec.Containers[0].Env) != 0 {
				var buffer bytes.Buffer
				for _, e := range f.Spec.Template.Spec.Containers[0].Env {
					buffer.WriteString(e.Name + " = " + e.Value + "\n")
				}
				env = buffer.String()
			}
			label := ""
			if len(f.Metadata.Labels) > 0 {
				var buffer bytes.Buffer
				for k, v := range f.Metadata.Labels {
					buffer.WriteString(k + " : " + v + "\n")
				}
				label = buffer.String()
			}
			table.AddRow(n, ns, h, r, t, tp, mem, env, label)
		}
		fmt.Fprintln(w, table)
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
