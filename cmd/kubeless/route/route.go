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
	"fmt"

	"github.com/kubeless/kubeless/pkg/spec"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
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

func getRouteDefinition(funcObj spec.Function, funcName, hostName, ns string, enableTLSAcme bool) (v1beta1.Ingress, error) {
	port := intstr.FromInt(8080)
	if len(funcObj.Spec.ServiceSpec.Ports) != 0 {
		port = funcObj.Spec.ServiceSpec.Ports[0].TargetPort
	}

	if port.IntVal <= 0 || port.IntVal > 65535 {
		return v1beta1.Ingress{}, fmt.Errorf("Invalid port number %d specified", port.IntVal)
	}

	ingress := v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      funcName,
			Namespace: ns,
			Labels:    funcObj.Metadata.Labels,
		},
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{
				{
					Host: hostName,
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{
							Paths: []v1beta1.HTTPIngressPath{
								{
									Path: "/",
									Backend: v1beta1.IngressBackend{
										ServiceName: funcObj.Metadata.Name,
										ServicePort: port,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	if enableTLSAcme {
		// add annotations and TLS configuration for kube-lego
		ingressAnnotations := map[string]string{
			"kubernetes.io/tls-acme":             "true",
			"ingress.kubernetes.io/ssl-redirect": "true",
		}
		ingress.ObjectMeta.Annotations = ingressAnnotations

		ingress.Spec.TLS = []v1beta1.IngressTLS{
			{
				Hosts:      []string{hostName},
				SecretName: funcName + "-tls",
			},
		}
	}

	return ingress, nil
}
