package main

import (
	"k8s.io/client-go/pkg/api/v1"
	"reflect"
	"testing"
)

func TestParseLabel(t *testing.T) {
	labels := []string{
		"foo=bar",
		"bar:foo",
		"foobar",
	}
	expected := map[string]string{
		"foo":    "bar",
		"bar":    "foo",
		"foobar": "",
	}
	actual := parseLabel(labels)
	if eq := reflect.DeepEqual(expected, actual); !eq {
		t.Errorf("Expect %v got %v", expected, actual)
	}
}

func TestParseEnv(t *testing.T) {
	envs := []string{
		"foo=bar",
		"bar:foo",
		"foobar",
		"foo=bar=baz",
	}
	expected := []v1.EnvVar{
		{
			Name:  "foo",
			Value: "bar",
		},
		{
			Name:  "bar",
			Value: "foo",
		},
		{
			Name:  "foobar",
			Value: "",
		},
		{
			Name:  "foo",
			Value: "bar=baz",
		},
	}
	actual := parseEnv(envs)
	if eq := reflect.DeepEqual(expected, actual); !eq {
		t.Errorf("Expect %v got %v", expected, actual)
	}
}
