package autoscale

import (
	"github.com/kubeless/kubeless/pkg/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var autoscaleCreateCmd = &cobra.Command{
	Use:   "create <name> FLAG",
	Short: "automatically scale function based on monitored metrics",
	Long:  `automatically scale function based on monitored metrics`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			logrus.Fatal("Need exactly one argument - function name")
		}
		funcName := args[0]

		ns, err := cmd.Flags().GetString("namespace")
		if err != nil {
			logrus.Fatal(err)
		}
		if ns == "" {
			ns = utils.GetDefaultNamespace()
		}

		function, err := utils.GetFunction(funcName, ns)
		if err != nil {
			logrus.Fatalf("Unable to find the function %s. Received %s: ", funcName, err)
		}

		min, err := cmd.Flags().GetInt32("min")
		if err != nil {
			logrus.Fatal(err)
		} else if min <= 0 {
			logrus.Fatalf("min can't be negative or zero")
		}
		max, err := cmd.Flags().GetInt32("max")
		if err != nil {
			logrus.Fatal(err)
		} else if max < min {
			logrus.Fatalf("max must be greater than or equal to min")
		}

		metric, err := cmd.Flags().GetString("metric")
		if err != nil {
			logrus.Fatal(err)
		}
		if metric != "cpu" && metric != "qps" {
			logrus.Fatalf("only supported metrics: cpu, qps")
		}

		value, err := cmd.Flags().GetString("value")
		if err != nil {
			logrus.Fatal(err)
		}

		hpa, err := getHorizontalAutoscaleDefinition(funcName, ns, metric, min, max, value, function.ObjectMeta.Labels)
		if err != nil {
			logrus.Fatal(err)
		}
		function.Spec.HorizontalPodAutoscaler = hpa

		kubelessClient, err := utils.GetKubelessClientOutCluster()
		if err != nil {
			logrus.Fatal(err)
		}
		logrus.Infof("Adding autoscaling rule to the function...")
		err = utils.UpdateFunctionCustomResource(kubelessClient, &function)
		if err != nil {
			logrus.Fatal(err)
		}
		logrus.Infof("Autoscaling rule for %s submitted for deployment", funcName)
	},
}

func init() {
	autoscaleCreateCmd.Flags().Int32("min", 1, "minimum number of replicas")
	autoscaleCreateCmd.Flags().Int32("max", 1, "maximum number of replicas")
	autoscaleCreateCmd.Flags().String("metric", "cpu", "metric to use for calculating the autoscale. Supported metrics: cpu, qps")
	autoscaleCreateCmd.Flags().String("value", "", "value of the average of the metric across all replicas. If metric is cpu, value is a number represented as percentage. If metric is qps, value must be in format of Quantity")
	autoscaleCreateCmd.MarkFlagRequired("value")
}
