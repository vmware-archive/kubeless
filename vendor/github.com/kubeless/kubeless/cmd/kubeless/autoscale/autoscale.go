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

package autoscale

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"k8s.io/api/autoscaling/v2beta1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//AutoscaleCmd contains first-class command for autoscale
var AutoscaleCmd = &cobra.Command{
	Use:   "autoscale SUBCOMMAND",
	Short: "manage autoscale to function on Kubeless",
	Long:  `autoscale command allows user to list, create, delete autoscale rule for function on Kubeless`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	cmds := []*cobra.Command{autoscaleCreateCmd, autoscaleListCmd, autoscaleDeleteCmd}

	for _, cmd := range cmds {
		AutoscaleCmd.AddCommand(cmd)
		cmd.Flags().StringP("namespace", "n", "", "Specify namespace for the autoscale")

	}
}

func getHorizontalAutoscaleDefinition(name, ns, metric string, min, max int32, value string, labels map[string]string) (v2beta1.HorizontalPodAutoscaler, error) {
	m := []v2beta1.MetricSpec{}
	switch metric {
	case "cpu":
		i, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return v2beta1.HorizontalPodAutoscaler{}, err
		}
		i32 := int32(i)
		m = []v2beta1.MetricSpec{
			{
				Type: v2beta1.ResourceMetricSourceType,
				Resource: &v2beta1.ResourceMetricSource{
					Name: v1.ResourceCPU,
					TargetAverageUtilization: &i32,
				},
			},
		}
	case "qps":
		q, err := resource.ParseQuantity(value)
		if err != nil {
			return v2beta1.HorizontalPodAutoscaler{}, err
		}
		m = []v2beta1.MetricSpec{
			{
				Type: v2beta1.ObjectMetricSourceType,
				Object: &v2beta1.ObjectMetricSource{
					MetricName:  "function_calls",
					TargetValue: q,
					Target: v2beta1.CrossVersionObjectReference{
						Kind: "Service",
						Name: name,
					},
				},
			},
		}
		if err != nil {
			return v2beta1.HorizontalPodAutoscaler{}, err
		}
	default:
		return v2beta1.HorizontalPodAutoscaler{}, fmt.Errorf("metric %s is not supported", metric)
	}

	return v2beta1.HorizontalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "autoscaling/v2beta1",
			Kind:       "HorizontalPodAutoscaler",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			Labels:    labels,
		},
		Spec: v2beta1.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: v2beta1.CrossVersionObjectReference{
				APIVersion: "apps/v1beta1",
				Kind:       "Deployment",
				Name:       name,
			},
			MinReplicas: &min,
			MaxReplicas: max,
			Metrics:     m,
		},
	}, nil
}
