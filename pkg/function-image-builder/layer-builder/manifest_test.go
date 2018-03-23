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
	"strings"
	"testing"
)

func TestNewManifest(t *testing.T) {
	manifestFile := strings.NewReader(`{"schemaVersion":2,"mediaType":"application/vnd.docker.distribution.manifest.v2+json","config":{"mediaType":"application/vnd.docker.container.image.v1+json","size":1489,"digest":"sha256:c7fc094ddbf9f9335543421b34d8c6f3becd3bb05c9f9a5ca0f0e6065871072d"},"layers":[{"mediaType":"application/vnd.docker.image.rootfs.diff.tar.gzip","size":723113,"digest":"sha256:d070b8ef96fc4f2d92ff520a4fe55594e362b4e1076a32bbfeb261dc03322910"}]}`)
	m := Manifest{}
	err := m.New(manifestFile)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	if m.Config.Size != 1489 {
		t.Errorf("Unexpected size %d", m.Config.Size)
	}
	if m.Config.Digest != "sha256:c7fc094ddbf9f9335543421b34d8c6f3becd3bb05c9f9a5ca0f0e6065871072d" {
		t.Errorf("Unexpected digest %s", m.Config.Digest)
	}
	if len(m.Layers) != 1 {
		t.Errorf("Unexpected layers length %d", len(m.Layers))
	}
}

func TestAddNewLayer(t *testing.T) {
	manifestFile := strings.NewReader(`{"schemaVersion":2,"mediaType":"application/vnd.docker.distribution.manifest.v2+json","config":{"mediaType":"application/vnd.docker.container.image.v1+json","size":1489,"digest":"sha256:c7fc094ddbf9f9335543421b34d8c6f3becd3bb05c9f9a5ca0f0e6065871072d"},"layers":[{"mediaType":"application/vnd.docker.image.rootfs.diff.tar.gzip","size":723113,"digest":"sha256:d070b8ef96fc4f2d92ff520a4fe55594e362b4e1076a32bbfeb261dc03322910"}]}`)
	m := Manifest{}
	err := m.New(manifestFile)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	m.AddLayer(&Layer{
		Size:   10,
		Sha256: "Test",
	})
	if len(m.Layers) != 2 {
		t.Errorf("Unexpected layers length %d", len(m.Layers))
	}
	if m.Layers[1].Size != 10 && m.Layers[1].Digest != "Test" {
		t.Errorf("Unexpected layer %v", m.Layers[1])
	}
}

func TestUpdateConfig(t *testing.T) {
	manifestFile := strings.NewReader(`{"schemaVersion":2,"mediaType":"application/vnd.docker.distribution.manifest.v2+json","config":{"mediaType":"application/vnd.docker.container.image.v1+json","size":1489,"digest":"sha256:c7fc094ddbf9f9335543421b34d8c6f3becd3bb05c9f9a5ca0f0e6065871072d"},"layers":[{"mediaType":"application/vnd.docker.image.rootfs.diff.tar.gzip","size":723113,"digest":"sha256:d070b8ef96fc4f2d92ff520a4fe55594e362b4e1076a32bbfeb261dc03322910"}]}`)
	m := Manifest{}
	err := m.New(manifestFile)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	m.UpdateConfig(&Layer{
		Size:   10,
		Sha256: "Test",
	})
	if m.Config.Size != 10 && m.Config.Digest != "Test" {
		t.Errorf("Unexpected layer %v", m.Config)
	}
}
