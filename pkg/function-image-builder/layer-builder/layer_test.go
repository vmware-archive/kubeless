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

package layerbuilder

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestNewLayer(t *testing.T) {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString("test content")
	layer := Layer{}
	err = layer.New(f)
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	if layer.Sha256 != "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" {
		t.Errorf("Wrong sha, expecting patata, received %s", layer.Sha256)
	}
	if layer.Size != 12 {
		t.Errorf("Wrong size, expecting patata, received %d", layer.Size)
	}
}
