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

package config

import (
	//"io/ioutil"
	//"os"
	//"testing"
)

var configStr = `
{
    "handler": {
        "slack": {
            "channel": "slack_channel",
            "token": "slack_token"
        }
    },
    "reason": ["created", "deleted", "updated"],
    "resource": {
    	"deployment": "false",
    	"replicationcontroller": "false",
    	"replicaset": "false",
    	"daemonset": "false",
    	"services": "false",
    	"pod": "false",
    },
}
`
//func TestLoadOK(t *testing.T) {
//	content := []byte(configStr)
//	tmpConfigFile, err := ioutil.TempFile(homeDir(), "kubeless")
//	if err != nil {
//		t.Fatalf("TestLoad(): %+v", err)
//	}
//
//	defer func() {
//		_ = os.Remove(tmpConfigFile.Name())
//	}()
//
//	if _, err := tmpConfigFile.Write(content); err != nil {
//		t.Fatalf("TestLoad(): %+v", err)
//	}
//	if err := tmpConfigFile.Close(); err != nil {
//		t.Fatalf("TestLoad(): %+v", err)
//	}
//
//	ConfigFileName = "kubeless"
//
//	c, err := New()
//
//	err = c.Load()
//	if err != nil {
//		t.Fatalf("TestLoad(): %+v", err)
//	}
//}