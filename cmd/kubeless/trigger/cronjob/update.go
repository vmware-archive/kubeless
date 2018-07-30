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

package cronjob

import (
	"fmt"

	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	cronjobUtils "github.com/kubeless/cronjob-trigger/pkg/utils"
	kubelessUtils "github.com/kubeless/kubeless/pkg/utils"
)

var updateCmd = &cobra.Command{
	Use:   "update <cronjob_trigger_name> FLAG",
	Short: "Update a cron job trigger",
	Long:  `Update a cron job trigger`,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) != 1 {
			logrus.Fatal("Need exactly one argument - cronjob trigger name")
		}
		triggerName := args[0]

		schedule, err := cmd.Flags().GetString("schedule")
		if err != nil {
			logrus.Fatal(err)
		}

		if schedule != "" {
			if _, err := cron.ParseStandard(schedule); err != nil {
				logrus.Fatalf("Invalid value for --schedule. " + err.Error())
			}
		}

		ns, err := cmd.Flags().GetString("namespace")
		if err != nil {
			logrus.Fatal(err)
		}
		if ns == "" {
			ns = kubelessUtils.GetDefaultNamespace()
		}

		functionName, err := cmd.Flags().GetString("function")
		if err != nil {
			logrus.Fatal(err)
		}

		dryrun, err := cmd.Flags().GetBool("dryrun")
		if err != nil {
			logrus.Fatal(err)
		}

		output, err := cmd.Flags().GetString("output")
		if err != nil {
			logrus.Fatal(err)
		}

		kubelessClient, err := kubelessUtils.GetKubelessClientOutCluster()
		if err != nil {
			logrus.Fatalf("Can not create out-of-cluster client: %v", err)
		}

		cronJobClient, err := cronjobUtils.GetKubelessClientOutCluster()
		if err != nil {
			logrus.Fatalf("Can not create out-of-cluster client: %v", err)
		}

		_, err = kubelessUtils.GetFunctionCustomResource(kubelessClient, functionName, ns)
		if err != nil {
			logrus.Fatalf("Unable to find Function %s in namespace %s. Error %s", triggerName, ns, err)
		}

		cronJobTrigger, err := cronjobUtils.GetCronJobCustomResource(cronJobClient, triggerName, ns)
		if err != nil {
			logrus.Fatalf("Unable to find Cronjob trigger %s in namespace %s. Error %s", triggerName, ns, err)
		}
		cronJobTrigger.Spec.FunctionName = functionName
		cronJobTrigger.Spec.Schedule = schedule

		if dryrun == true {
			res, err := kubelessUtils.DryRunFmt(output, cronJobTrigger)
			if err != nil {
				logrus.Fatal(err)
			}
			fmt.Println(res)
			return
		}

		err = cronjobUtils.UpdateCronJobCustomResource(cronJobClient, cronJobTrigger)
		if err != nil {
			logrus.Fatalf("Failed to update cronjob trigger object %s in namespace %s. Error: %s", triggerName, ns, err)
		}
		logrus.Infof("Cronjob trigger %s updated in namespace %s successfully!", triggerName, ns)
	},
}

func init() {
	updateCmd.Flags().StringP("namespace", "n", "", "Specify namespace of the cronjob trigger")
	updateCmd.Flags().StringP("schedule", "", "", "Specify schedule in cron format for scheduled function")
	updateCmd.Flags().StringP("function", "", "", "Name of the function to be associated with trigger")
	updateCmd.Flags().Bool("dryrun", false, "Output JSON manifest of the function without creating it")
	updateCmd.Flags().StringP("output", "o", "yaml", "Output format")
}
