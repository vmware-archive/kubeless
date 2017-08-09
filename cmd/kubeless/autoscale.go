package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/kubeless/kubeless/pkg/utils"
	"github.com/spf13/cobra"
)

var autoscaleCmd = &cobra.Command{
	Use:   "autoscale SUBCOMMAND",
	Short: "automatically scale function based on monitored metrics",
	Long:  `automatically scale function based on monitored metrics`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			logrus.Fatal("Need exactly one argument - function name")
		}
		funcName := args[0]

		min, err := cmd.Flags().GetInt32("min")
		if err != nil {
			logrus.Fatal(err)
		} else if min == 0 {
			logrus.Fatalf("--min can't be 0")
		}
		max, err := cmd.Flags().GetInt32("max")
		if err != nil {
			logrus.Fatal(err)
		} else if min == 0 {
			logrus.Fatalf("--max can't be 0")
		}

		client := utils.GetClientOutOfCluster()

		err = utils.CreateAutoscale(client, funcName, min, max)
		if err != nil {
			logrus.Fatalf("Can't create autoscale: %v", err)
		}
	},
}

func init() {
	autoscaleCmd.Flags().Int32("min", 0, "minimum number of replicas")
	autoscaleCmd.Flags().Int32("max", 0, "maximum number of replicas")
}
