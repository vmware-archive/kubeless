package registry

import (
	"reflect"
	"testing"

	"k8s.io/api/core/v1"
)

func TestNew(t *testing.T) {
	s := v1.Secret{
		Data: map[string][]byte{
			".dockerconfigjson": []byte("{\"auths\":{\"https://index.docker.io/v1/\":{\"username\":\"test\",\"password\":\"pass\"}}}"),
		},
	}
	r, err := New(s)
	if err != nil {
		t.Error(err)
	}
	if r.Endpoint != "https://index.docker.io" {
		t.Errorf("Unexpected endpoint %s, expecting https://index.docker.io", r.Endpoint)
	}
	if r.Version != "v1" {
		t.Errorf("Unexpected version %s, expecting v1", r.Version)
	}
	if r.Creds.Username != "test" {
		t.Errorf("Unexpected username %s, expecting test", r.Creds.Username)
	}
	if r.Creds.Password != "pass" {
		t.Errorf("Unexpected password %s, expecting pass", r.Creds.Password)
	}
}

func TestTagURLV1(t *testing.T) {
	r := Registry{
		Endpoint: "https://registry-1.docker.io",
		Version:  "v1",
	}
	url, err := r.tagURL("test/image")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if url != "https://registry-1.docker.io/v1/repositories/test/image/tags" {
		t.Errorf("Unexpected URL %s", url)
	}
}

func TestTagURLV2(t *testing.T) {
	r := Registry{
		Endpoint: "https://registry-1.docker.io",
		Version:  "v2",
	}
	url, err := r.tagURL("test/image")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if url != "https://registry-1.docker.io/v2/test/image/tags/list" {
		t.Errorf("Unexpected URL %s", url)
	}
}

func TestGetTagsV1(t *testing.T) {
	r := Registry{
		Endpoint: "https://registry-1.docker.io",
		Version:  "v1",
	}
	body := []byte("[{\"later\": \"\", \"name\": \"latest\"}]")
	tags, err := r.getTags(body)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	expectedTags := []string{"latest"}
	if !reflect.DeepEqual(tags, expectedTags) {
		t.Errorf("Unexpected tags: %v", tags)
	}
}

func TestGetTagsV2(t *testing.T) {
	r := Registry{
		Endpoint: "https://registry-1.docker.io",
		Version:  "v2",
	}
	body := []byte("{\"name\": \"test\", \"tags\":[\"latest\"]}")
	tags, err := r.getTags(body)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	expectedTags := []string{"latest"}
	if !reflect.DeepEqual(tags, expectedTags) {
		t.Errorf("Unexpected tags: %v", tags)
	}
}
