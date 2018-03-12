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
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
)

type layer struct {
	MediaType string `json:"mediaType"`
	Size      int64  `json:"size"`
	Digest    string `json:"digest"`
}

// Manifest represent the manifest.json of an image
type Manifest struct {
	SchemaVersion int     `json:"schemaVersion"`
	MediaType     string  `json:"mediaType"`
	Config        layer   `json:"config"`
	Layers        []layer `json:"layers"`
}

// New parses an io.Reader into a Manifest
func (m *Manifest) New(manifestFile io.Reader) error {
	manifestContent, err := ioutil.ReadAll(manifestFile)
	if err != nil {
		return err
	}
	err = json.Unmarshal(manifestContent, m)
	if err != nil {
		return nil
	}
	return nil
}

// UpdateConfig overrides the Config information of the manifest with a new Layer
func (m *Manifest) UpdateConfig(newConfig *Layer) {
	m.Config.Size = int64(newConfig.Size)
	m.Config.Digest = fmt.Sprintf("sha256:%s", newConfig.Sha256)
}

// AddLayer adds a new layer to the list in the Manifest
func (m *Manifest) AddLayer(newLayer *Layer) {
	m.Layers = append(m.Layers, layer{
		MediaType: "application/vnd.docker.image.rootfs.diff.tar.gzip",
		Size:      newLayer.Size,
		Digest:    fmt.Sprintf("sha256:%s", newLayer.Sha256),
	})
}
