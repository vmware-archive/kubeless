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

package completion

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	completionShells = map[string]func(out io.Writer, cmd *cobra.Command) error{
		"bash": runCompletionBash,
	}
)

var bashCmd = &cobra.Command{

	Use:   "bash",
	Short: "output shell completion code for bash",
	Long:  `output shell completion code for bash`,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) > 0 {
			logrus.Fatalf("Too many arguments. Expected only the shell type.")
		}

		run, found := completionShells[cmd.Name()]
		if !found {
			logrus.Fatalf("Unsupported shell type.")
		}

		run(os.Stdout, cmd.Parent().Parent())
	},
}

func runCompletionBash(out io.Writer, cmd *cobra.Command) error {
	return cmd.GenBashCompletion(out)
}
