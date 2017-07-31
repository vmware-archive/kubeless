package main

import (
	"bytes"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	xv1beta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

func listIngressOutput(t *testing.T, client kubernetes.Interface, ns, output string) string {
	var buf bytes.Buffer

	if err := doIngressList(&buf, client, ns, output); err != nil {
		t.Fatalf("doList returned error: %v", err)
	}

	return buf.String()
}

func TestIngressList(t *testing.T) {
	ing1 := xv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "myns",
		},
		Spec: xv1beta1.IngressSpec{
			Rules: []xv1beta1.IngressRule{
				{
					Host: "foobar.192.168.99.100.nip.io",
					IngressRuleValue: xv1beta1.IngressRuleValue{
						HTTP: &xv1beta1.HTTPIngressRuleValue{
							Paths: []xv1beta1.HTTPIngressPath{
								{
									Path: "/",
									Backend: xv1beta1.IngressBackend{
										ServiceName: "foobar",
										ServicePort: intstr.FromInt(8080),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	ing2 := xv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bar",
			Namespace: "myns",
		},
		Spec: xv1beta1.IngressSpec{
			Rules: []xv1beta1.IngressRule{
				{
					Host: "example.com",
					IngressRuleValue: xv1beta1.IngressRuleValue{
						HTTP: &xv1beta1.HTTPIngressRuleValue{
							Paths: []xv1beta1.HTTPIngressPath{
								{
									Path: "/",
									Backend: xv1beta1.IngressBackend{
										ServiceName: "barfoo",
										ServicePort: intstr.FromInt(8080),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	client := fake.NewSimpleClientset(&ing1, &ing2)

	output := listIngressOutput(t, client, "myns", "")
	t.Log("output is", output)

	if !strings.Contains(output, "foo") || !strings.Contains(output, "bar") {
		t.Errorf("table output didn't mention both functions")
	}

	// json output
	output = listIngressOutput(t, client, "myns", "json")
	t.Log("output is", output)

	// yaml output
	output = listIngressOutput(t, client, "myns", "yaml")
	t.Log("output is", output)
}
