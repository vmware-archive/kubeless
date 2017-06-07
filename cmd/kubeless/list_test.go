package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/rest/fake"

	"github.com/kubeless/kubeless/pkg/spec"
)

func fakeTprClient(f func(req *http.Request) (*http.Response, error)) *fake.RESTClient {
	return &fake.RESTClient{
		APIRegistry:          api.Registry,
		NegotiatedSerializer: api.Codecs,
		Client:               fake.CreateHTTPClient(f),
	}
}

func listOutput(t *testing.T, client rest.Interface, ns, output string, args []string) string {
	var buf bytes.Buffer

	if err := doList(&buf, client, ns, output, args); err != nil {
		t.Fatalf("doList returned error: %v", err)
	}

	return buf.String()
}

func objBody(object interface{}) io.ReadCloser {
	output, err := json.Marshal(object)
	if err != nil {
		panic(err)
	}
	return ioutil.NopCloser(bytes.NewReader([]byte(output)))
}

func TestList(t *testing.T) {
	listObj := spec.FunctionList{
		Items: []*spec.Function{
			{
				Metadata: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "myns",
				},
				Spec: spec.FunctionSpec{
					Handler:  "fhandler",
					Function: "ffunction",
					Runtime:  "fruntime",
					Type:     "ftype",
					Topic:    "ftopic",
					Deps:     "fdeps",
				},
			},
			{
				Metadata: metav1.ObjectMeta{
					Name:      "bar",
					Namespace: "myns",
				},
				Spec: spec.FunctionSpec{
					Handler:  "bhandler",
					Function: "bfunction",
					Runtime:  "bruntime",
					Type:     "btype",
					Topic:    "btopic",
					Deps:     "bdeps",
				},
			},
		},
	}

	client := fakeTprClient(func(req *http.Request) (*http.Response, error) {
		header := http.Header{}
		header.Set("Content-Type", runtime.ContentTypeJSON)
		switch req.URL.Path {
		case "/namespaces/myns/functions":
			return &http.Response{
				StatusCode: 200,
				Header:     header,
				Body:       objBody(&listObj),
			}, nil
		case "/namespaces/myns/functions/foo":
			return &http.Response{
				StatusCode: 200,
				Header:     header,
				Body:       objBody(listObj.Items[0]),
			}, nil
		default:
			t.Fatalf("unexpected request: %#v\n%#v", req.URL, req)
			return nil, nil
		}
	})

	// No arg -> list everything in namespace
	output := listOutput(t, client, "myns", "", []string{})
	t.Log("output is", output)

	if !strings.Contains(output, "foo") || !strings.Contains(output, "bar") {
		t.Errorf("table output didn't mention both functions")
	}

	// Explicit arg(s)
	output = listOutput(t, client, "myns", "", []string{"foo"})
	t.Log("output is", output)

	if !strings.Contains(output, "foo") {
		t.Errorf("table output didn't mention explicit function foo")
	}
	if strings.Contains(output, "bar") {
		t.Errorf("table output mentions unrequested function bar")
	}

	// TODO: Actually validate the output of the following.
	// Probably need to fix output framing first.

	// json output
	output = listOutput(t, client, "myns", "json", []string{})
	t.Log("output is", output)

	// yaml output
	output = listOutput(t, client, "myns", "yaml", []string{})
	t.Log("output is", output)
}
