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

package route

import (
	"github.com/spf13/cobra"
)

//RouteCmd contains first-class command for route
var RouteCmd = &cobra.Command{
	Use:   "route SUBCOMMAND",
	Short: "manage route to function on Kubeless",
	Long:  `route command allows user to list, create, delete routing rule for function on Kubeless`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	cmds := []*cobra.Command{routeCreateCmd, routeListCmd, routeDeleteCmd}

	for _, cmd := range cmds {
		RouteCmd.AddCommand(cmd)
		cmd.Flags().StringP("namespace", "n", "", "Specify namespace for the route")

	}
}
