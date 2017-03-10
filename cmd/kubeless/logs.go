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
	"io"
	"os"
		
	"github.com/Sirupsen/logrus"
	"github.com/skippbox/kubeless/pkg/utils"
	"github.com/spf13/cobra"
	"k8s.io/kubernetes/pkg/api"
)

var logsCmd = &cobra.Command{
	Use:   "logs <function_name> FLAG",
	Short: "get logs from a running function",
	Long:  `get logs from a running function`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			logrus.Fatal("Need exactly one argument - function name")
		}
		funcName := args[0]
		follow, err := cmd.Flags().GetBool("follow")
		if err != nil {
			logrus.Fatal(err)
		}
		ns, err := cmd.Flags().GetString("namespace")
		if err != nil {
			logrus.Fatal(err)
		}

		f := utils.GetFactory()
		if ns == "" {
			ns, _, err = f.DefaultNamespace()
			if err != nil {
				logrus.Fatalf("Getting log failed: %v", err)
			}
		}
		kClient, err := f.Client()
		if err != nil {
			logrus.Fatalf("Getting log failed: %v", err)
		}
		podName, err := utils.GetPodName(kClient, ns, funcName)
		podLog := &api.PodLogOptions{
			Container: funcName,
			Follow:    follow,
		}
		req := kClient.Pods(ns).GetLogs(podName, podLog)

		readCloser, err := req.Stream()
		if err != nil {
			logrus.Fatalf("Getting log failed: %v", err)
		}
		defer readCloser.Close()
		io.Copy(os.Stdout, readCloser)
	},
}

func init() {
	logsCmd.Flags().BoolP("follow", "f", false, "Specify if the logs should be streamed.")
	logsCmd.Flags().StringP("namespace", "", "", "Specify namespace for the function")
}
