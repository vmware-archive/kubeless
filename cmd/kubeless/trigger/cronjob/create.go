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
	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	kubelessApi "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	"github.com/kubeless/kubeless/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var createCmd = &cobra.Command{
	Use:   "create <cronjob_trigger_name> FLAG",
	Short: "Create a cron job trigger",
	Long:  `Create a cron job trigger`,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) != 1 {
			logrus.Fatal("Need exactly one argument - cronjob trigger name")
		}
		triggerName := args[0]

		schedule, err := cmd.Flags().GetString("schedule")
		if err != nil {
			logrus.Fatal(err)
		}

		if _, err := cron.ParseStandard(schedule); err != nil {
			logrus.Fatalf("Invalid value for --schedule. " + err.Error())
		}

		ns, err := cmd.Flags().GetString("namespace")
		if err != nil {
			logrus.Fatal(err)
		}
		if ns == "" {
			ns = utils.GetDefaultNamespace()
		}

		functionName, err := cmd.Flags().GetString("function")
		if err != nil {
			logrus.Fatal(err)
		}

		kubelessClient, err := utils.GetKubelessClientOutCluster()
		if err != nil {
			logrus.Fatalf("Can not create out-of-cluster client: %v", err)
		}

		_, err = utils.GetFunctionCustomResource(kubelessClient, functionName, ns)
		if err != nil {
			logrus.Fatalf("Unable to find Function %s in namespace %s. Error %s", functionName, ns, err)
		}

		cronJobTrigger := kubelessApi.CronJobTrigger{}
		cronJobTrigger.TypeMeta = metav1.TypeMeta{
			Kind:       "CronJobTrigger",
			APIVersion: "kubeless.io/v1beta1",
		}
		cronJobTrigger.ObjectMeta = metav1.ObjectMeta{
			Name:      triggerName,
			Namespace: ns,
		}
		cronJobTrigger.ObjectMeta.Labels = map[string]string{
			"created-by": "kubeless",
		}
		cronJobTrigger.Spec.FunctionName = functionName
		cronJobTrigger.Spec.Schedule = schedule
		err = utils.CreateCronJobCustomResource(kubelessClient, &cronJobTrigger)
		if err != nil {
			logrus.Fatalf("Failed to create cronjob trigger object %s in namespace %s. Error: %s", triggerName, ns, err)
		}
		logrus.Infof("Cronjob trigger %s created in namespace %s successfully!", triggerName, ns)
	},
}

func init() {
	createCmd.Flags().StringP("namespace", "", "", "Specify namespace for the cronjob trigger")
	createCmd.Flags().StringP("schedule", "", "", "Specify schedule in cron format for scheduled function")
	createCmd.Flags().StringP("function", "", "", "Name of the function to be associated with trigger")
	createCmd.MarkFlagRequired("function")
	createCmd.MarkFlagRequired("schedule")
}
