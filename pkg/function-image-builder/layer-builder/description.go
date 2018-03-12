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
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"time"
)

// Config represents a container configuration
type Config struct {
	Hostname     string
	Domainname   string
	User         string
	AttachStdin  bool
	AttachStdout bool
	AttachStderr bool
	Tty          bool
	OpenStdin    bool
	StdinOnce    bool
	Env          []string
	Cmd          []string
	ArgsEscaped  bool
	Image        string
	Volumes      interface{}
	WorkingDir   string
	Entrypoint   interface{}
	OnBuild      interface{}
	Labels       interface{}
}

// HistoryEntry represents a layer creation info
type HistoryEntry struct {
	Created    string `json:"created"`
	CreatedBy  string `json:"created_by,omitifempty"`
	Comment    string `json:"comment,omitifempty"`
	EmptyLayer bool   `json:"empty_layer,omitifempty"`
}

// Rootfs represents the root filesystem of an image
type Rootfs struct {
	Type    string   `json:"type"`
	DiffIds []string `json:"diff_ids"`
}

// Description represents the specification of a Docker image
type Description struct {
	Arch            string         `json:"architecture"`
	Config          Config         `json:"config"`
	Container       string         `json:"container"`
	ContainerConfig Config         `json:"container_config"`
	Created         string         `json:"created"`
	DockerVersion   string         `json:"docker_version"`
	History         []HistoryEntry `json:"history"`
	OS              string         `json:"os"`
	Rootfs          Rootfs         `json:"rootfs"`
}

// New generates a Description object based on the description file
func (d *Description) New(descriptionFile io.Reader) error {
	descriptionContent, err := ioutil.ReadAll(descriptionFile)
	if err != nil {
		return err
	}
	return json.Unmarshal(descriptionContent, d)
}

// AddLayer adds a new Layer to the image Description
func (d *Description) AddLayer(newLayer *Layer) {
	//   Delete some properties that doesn't apply anymore
	d.Config.Hostname = ""
	d.Config.Image = ""
	d.Container = ""
	d.ContainerConfig.Hostname = ""
	d.ContainerConfig.Image = ""
	//   Update new properties
	d.Created = time.Now().UTC().Format(time.RFC3339)
	d.History = append(d.History, HistoryEntry{
		Created: time.Now().UTC().Format(time.RFC3339),
		Comment: "Created by Kubeless",
	})
	d.Rootfs.DiffIds = append(d.Rootfs.DiffIds, fmt.Sprintf("sha256:%s", newLayer.Sha256))
}

// Content returns the description content
func (d *Description) Content() ([]byte, error) {
	return json.Marshal(*d)
}

// ToLayer returns the Description as a Layer
func (d *Description) ToLayer() (*Layer, error) {
	content, err := d.Content()
	if err != nil {
		return nil, err
	}
	descriptionNewSize := int64(len(content))
	descriptionNewSha := fmt.Sprintf("%x", sha256.Sum256(content))

	return &Layer{
		Size:   descriptionNewSize,
		Sha256: descriptionNewSha,
	}, nil
}
