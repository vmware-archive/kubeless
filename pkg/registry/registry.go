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

package registry

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"regexp"
	"time"

	"k8s.io/api/core/v1"
)

// Credentials represent the required credentials to authenticate against a Docker registry
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email,omitifempty"`
	Auth     string `json:"auth,omitifempty"`
}

// Registry struct represents a Docker Registry
type Registry struct {
	Endpoint string
	Version  string
	Creds    Credentials
}

type tagv1 struct {
	Layer string `json:"layer"`
	Name  string `json:"name"`
}

type tagListV2 struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

type dockerCfg struct {
	Auths map[string]Credentials `json:"auths"`
}

// New returns a Registry struct parsing its URL and storing the required credentials
func New(config v1.Secret) (*Registry, error) {
	// Parse secret
	cfg := dockerCfg{}
	err := json.Unmarshal(config.Data[".dockerconfigjson"], &cfg)
	if err != nil {
		return nil, err
	}
	regs := reflect.ValueOf(cfg.Auths).MapKeys()
	if len(regs) > 1 {
		return nil, fmt.Errorf("Found several registries: %q, unable to decide which one to use", regs)
	}
	registryURL := regs[0].String()
	re := regexp.MustCompile("(https?://.*)/(v[0-9]+)/?")
	parsedURL := re.FindStringSubmatch(registryURL)
	if len(parsedURL) == 0 {
		return nil, fmt.Errorf("Unable to parse registry URL %s", registryURL)
	}
	reg := Registry{
		Endpoint: parsedURL[1],
		Version:  parsedURL[2],
		Creds:    cfg.Auths[registryURL],
	}
	return &reg, err
}

// getTags return the list of tags from an HTTP response to the tag/list API endpoint
func (r *Registry) getTags(body []byte) ([]string, error) {
	switch r.Version {
	case "v1":
		response := []tagv1{}
		err := json.Unmarshal(body, &response)
		if err != nil {
			return nil, err
		}
		tags := []string{}
		for _, tag := range response {
			tags = append(tags, tag.Name)
		}
		return tags, nil
	case "v2":
		response := tagListV2{}
		err := json.Unmarshal(body, &response)
		if err != nil {
			return nil, err
		}
		return response.Tags, nil
	default:
		return nil, fmt.Errorf("API version %s not supported", r.Version)
	}
}

// tagURL return the URL of the endpoint for listing existing tags
func (r *Registry) tagURL(img string) (string, error) {
	switch r.Version {
	case "v1":
		return fmt.Sprintf("%s/%s/repositories/%s/tags", r.Endpoint, r.Version, img), nil
	case "v2":
		return fmt.Sprintf("%s/%s/%s/tags/list", r.Endpoint, r.Version, img), nil
	default:
		return "", fmt.Errorf("API version %s not supported", r.Version)
	}
}

// findProperty returns the value of a property from a list witht the format 'foo="bar",bar="foo"'
func findProperty(src, property string) (string, error) {
	re := regexp.MustCompile(fmt.Sprintf("%s=\"([^\"]*)\"", property))
	res := re.FindStringSubmatch(src)
	if len(res) != 2 {
		return "", fmt.Errorf("Unable to find the property %s in %s", property, src)
	}
	return res[1], nil
}

type authResponse struct {
	Token string `json:"token"`
}

// doRequestWithAuth does an HTTP GET agains the given url parsing the authInfo given
func doRequestWithAuth(authInfo, url string, client *http.Client) ([]byte, error) {
	bearer, err := findProperty(authInfo, "Bearer realm")
	if err != nil {
		return nil, fmt.Errorf("Unable to extract auth info: %v", err)
	}
	service, err := findProperty(authInfo, "service")
	if err != nil {
		return nil, fmt.Errorf("Unable to extract auth info: %v", err)
	}
	scope, err := findProperty(authInfo, "scope")
	if err != nil {
		return nil, fmt.Errorf("Unable to extract auth info: %v", err)
	}
	authResp, err := client.Get(fmt.Sprintf("%s?service=%s&scope=%s", bearer, service, scope))
	if err != nil {
		return nil, fmt.Errorf("Unable to obtain auth token: %v", err)
	}
	defer authResp.Body.Close()
	authb, err := ioutil.ReadAll(authResp.Body)
	if err != nil {
		return nil, err
	}
	authr := authResponse{}
	err = json.Unmarshal(authb, &authr)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse auth token: %v", err)
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", authr.Token))
	respWithAuth, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer respWithAuth.Body.Close()
	body, err := ioutil.ReadAll(respWithAuth.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (r *Registry) doRequest(url string) ([]byte, error) {
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	client := &http.Client{
		Transport: tr,
	}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// Handle auth if needed
	if resp.StatusCode == 401 {
		// Get auth info from headers
		authInfo := resp.Header.Get("Www-Authenticate")
		if authInfo == "" {
			return nil, fmt.Errorf("Failed to authenticate: unknown authentication format: %v", body)
		}
		body, err = doRequestWithAuth(authInfo, url, client)
		if err != nil {
			return nil, err
		}
	}
	return body, nil
}

// ImageExists checks if a certain image:tag exists in the registry
func (r *Registry) ImageExists(id, tag string) (bool, error) {
	url, err := r.tagURL(id)
	if err != nil {
		return false, err
	}
	body, err := r.doRequest(url)
	if err != nil {
		return false, err
	}
	if match, _ := regexp.MatchString("Resource not found", string(body)); match {
		// There is no image with that ID yet
		return false, nil
	}
	tags, err := r.getTags(body)
	if err != nil {
		return false, err
	}
	for _, t := range tags {
		if t == tag {
			return true, nil
		}
	}
	return false, nil
}
