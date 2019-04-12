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
	"github.com/spf13/cobra"
)

//CompletionCmd contains first-class command for completion
var CompletionCmd = &cobra.Command{
	Use:   "completion [shell]",
	Short: "Output shell completion code for the specified shell.",
	Long: `Output shell completion code for the specified shell. For bash, load the completion code into the current shell: 
	
	source <(kubeless completion bash)`,
}

func init() {
	CompletionCmd.AddCommand(bashCmd)
}
