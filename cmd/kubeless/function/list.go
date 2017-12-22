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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/gosuri/uitable"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
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
		if ns == "" {
			ns = utils.GetDefaultNamespace()
		}

		crdClient, err := utils.GetCRDClientOutOfCluster()
		if err != nil {
			logrus.Fatalf("Can not list functions: %v", err)
		}

		apiV1Client := utils.GetClientOutOfCluster()

		if err := doList(cmd.OutOrStdout(), crdClient, apiV1Client, ns, output, args); err != nil {
			logrus.Fatal(err.Error())
		}
	},
}

func init() {
	listCmd.Flags().StringP("out", "o", "", "Output format. One of: json|yaml")
	listCmd.Flags().StringP("namespace", "n", "", "Specify namespace for the function")
}

func doList(w io.Writer, crdClient rest.Interface, apiV1Client kubernetes.Interface, ns, output string, args []string) error {
	var list []*spec.Function
	if len(args) == 0 {
		funcList := spec.FunctionList{}
		err := crdClient.Get().
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
			err := crdClient.Get().
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

	return printFunctions(w, list, apiV1Client, output)
}

func parseDeps(deps, runtime string) (res string, err error) {
	if deps != "" {
		if strings.Contains(runtime, "nodejs") {
			pkgjson := make(map[string]interface{})
			err = json.Unmarshal([]byte(deps), &pkgjson)
			if err != nil {
				return
			}
			if pkgjson["dependencies"] != nil {
				dependencies := []string{}
				for pkg, ver := range pkgjson["dependencies"].(map[string]interface{}) {
					dependencies = append(dependencies, pkg+": "+ver.(string))
				}
				res = strings.Join(dependencies, "\n")
			}
		} else {
			res = deps
		}
	}
	return
}

// printFunctions formats the output of function list
func printFunctions(w io.Writer, functions []*spec.Function, cli kubernetes.Interface, output string) error {
	if output == "" {
		table := uitable.New()
		table.MaxColWidth = 50
		table.Wrap = true
		table.AddRow("NAME", "NAMESPACE", "HANDLER", "RUNTIME", "TYPE", "TOPIC", "DEPENDENCIES", "STATUS")
		for _, f := range functions {
			n := f.Metadata.Name
			h := f.Spec.Handler
			r := f.Spec.Runtime
			t := f.Spec.Type
			tp := f.Spec.Topic
			ns := f.Metadata.Namespace
			status, err := getDeploymentStatus(cli, f.Metadata.Name, f.Metadata.Namespace)
			if err != nil && k8sErrors.IsNotFound(err) {
				status = "MISSING: Check controller logs"
			} else if err != nil {
				return err
			}
			deps, err := parseDeps(f.Spec.Deps, r)
			if err != nil {
				return err
			}
			table.AddRow(n, ns, h, r, t, tp, deps, status)
		}
		fmt.Fprintln(w, table)
	} else if output == "wide" {
		table := uitable.New()
		table.MaxColWidth = 50
		table.Wrap = true
		table.AddRow("NAME", "NAMESPACE", "HANDLER", "RUNTIME", "TYPE", "TOPIC", "DEPENDENCIES", "STATUS", "MEMORY", "ENV", "LABEL", "SCHEDULE")
		for _, f := range functions {
			n := f.Metadata.Name
			h := f.Spec.Handler
			r := f.Spec.Runtime
			t := f.Spec.Type
			tp := f.Spec.Topic
			deps, err := parseDeps(f.Spec.Deps, r)
			if err != nil {
				return err
			}
			s := f.Spec.Schedule
			ns := f.Metadata.Namespace
			status, err := getDeploymentStatus(cli, f.Metadata.Name, f.Metadata.Namespace)
			if err != nil && k8sErrors.IsNotFound(err) {
				status = "MISSING: Check controller logs"
			} else if err != nil {
				return err
			}
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
			table.AddRow(n, ns, h, r, t, tp, deps, status, mem, env, label, s)
		}
		fmt.Fprintln(w, table)
	} else {
		for _, f := range functions {
			switch output {
			case "json":
				b, err := json.MarshalIndent(f, "", "  ")
				if err != nil {
					return err
				}
				fmt.Fprintln(w, string(b))
			case "yaml":
				b, err := yaml.Marshal(f)
				if err != nil {
					return err
				}
				fmt.Fprintln(w, string(b))
			default:
				return fmt.Errorf("Wrong output format. Please use only json|yaml")
			}
		}
	}
	return nil
}
