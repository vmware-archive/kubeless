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

package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	kubelessutil "github.com/kubeless/kubeless/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// IsJSON returns true if the string is json
func IsJSON(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil

}

// GetHTTPReq returns the http request object that can be used to send a event with payload to function service
func GetHTTPReq(clientset kubernetes.Interface, funcName, namespace, eventNamespace, method, body string) (*http.Request, error) {
	svc, err := clientset.CoreV1().Services(namespace).Get(funcName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("Unable to find the service for function %s", funcName)
	}
	funcPort := strconv.Itoa(int(svc.Spec.Ports[0].Port))
	req, err := http.NewRequest(method, fmt.Sprintf("http://%s.%s.svc.cluster.local:%s", funcName, namespace, funcPort), strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("Unable to create request %v", err)
	}
	timestamp := time.Now().UTC()
	eventID, err := kubelessutil.GetRandString(11)
	if err != nil {
		return nil, fmt.Errorf("Failed to create a event-ID %v", err)
	}
	req.Header.Add("event-id", eventID)
	req.Header.Add("event-time", timestamp.String())
	req.Header.Add("event-namespace", eventNamespace)
	if IsJSON(body) {
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("event-type", "application/json")
	} else {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("event-type", "application/x-www-form-urlencoded")
	}
	return req, nil
}

// SendMessage sends messge over function service
func SendMessage(req *http.Request) error {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error: received error code %d: %s", resp.StatusCode, resp.Status)
	}
	return nil
}
