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

func TestNewDescription(t *testing.T) {
	descFile := strings.NewReader(`{"architecture":"amd64","config":{"Hostname":"","Domainname":"","User":"","AttachStdin":false,"AttachStdout":false,"AttachStderr":false,"Tty":false,"OpenStdin":false,"StdinOnce":false,"Env":["PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"],"Cmd":["sh"],"ArgsEscaped":true,"Image":"sha256:8cae5980d887cc55ba2f978ae99c662007ee06d79881678d57f33f0473fe0736","Volumes":null,"WorkingDir":"","Entrypoint":null,"OnBuild":null,"Labels":null},"container":"8d2c840a1a9b2544fe713c2e24b6757d52328f09bdfc9c2ef6219afbf7ae6b59","container_config":{"Hostname":"8d2c840a1a9b","Domainname":"","User":"","AttachStdin":false,"AttachStdout":false,"AttachStderr":false,"Tty":false,"OpenStdin":false,"StdinOnce":false,"Env":["PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"],"Cmd":["/bin/sh","-c","#(nop) "],"ArgsEscaped":true,"Image":"sha256:8cae5980d887cc55ba2f978ae99c662007ee06d79881678d57f33f0473fe0736","Volumes":null,"WorkingDir":"","Entrypoint":null,"OnBuild":null,"Labels":{}},"created":"2018-02-28T22:14:49.023807051Z","docker_version":"17.06.2-ce","history":[{"created":"2018-02-28T22:14:48.759033366Z","created_by":"/bin/sh -c #(nop) ADD file:327f69fc1ac9a7b6e56e9032f7b8fbd7741dd0b22920761909c6c8e5fa9c5815 in / "},{"created":"2018-02-28T22:14:49.023807051Z","created_by":"/bin/sh -c #(nop)  ","empty_layer":true}],"os":"linux","rootfs":{"type":"layers","diff_ids":["sha256:c5183829c43c4698634093dc38f9bee26d1b931dedeba71dbee984f42fe1270d"]}}`)
	d := Description{}
	err := d.New(descFile)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
}

func TestAddLayerDescription(t *testing.T) {
	descFile := strings.NewReader(`{"architecture":"amd64","config":{"Hostname":"","Domainname":"","User":"","AttachStdin":false,"AttachStdout":false,"AttachStderr":false,"Tty":false,"OpenStdin":false,"StdinOnce":false,"Env":["PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"],"Cmd":["sh"],"ArgsEscaped":true,"Image":"sha256:8cae5980d887cc55ba2f978ae99c662007ee06d79881678d57f33f0473fe0736","Volumes":null,"WorkingDir":"","Entrypoint":null,"OnBuild":null,"Labels":null},"container":"8d2c840a1a9b2544fe713c2e24b6757d52328f09bdfc9c2ef6219afbf7ae6b59","container_config":{"Hostname":"8d2c840a1a9b","Domainname":"","User":"","AttachStdin":false,"AttachStdout":false,"AttachStderr":false,"Tty":false,"OpenStdin":false,"StdinOnce":false,"Env":["PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"],"Cmd":["/bin/sh","-c","#(nop) "],"ArgsEscaped":true,"Image":"sha256:8cae5980d887cc55ba2f978ae99c662007ee06d79881678d57f33f0473fe0736","Volumes":null,"WorkingDir":"","Entrypoint":null,"OnBuild":null,"Labels":{}},"created":"2018-02-28T22:14:49.023807051Z","docker_version":"17.06.2-ce","history":[{"created":"2018-02-28T22:14:48.759033366Z","created_by":"/bin/sh -c #(nop) ADD file:327f69fc1ac9a7b6e56e9032f7b8fbd7741dd0b22920761909c6c8e5fa9c5815 in / "},{"created":"2018-02-28T22:14:49.023807051Z","created_by":"/bin/sh -c #(nop)  ","empty_layer":true}],"os":"linux","rootfs":{"type":"layers","diff_ids":["sha256:c5183829c43c4698634093dc38f9bee26d1b931dedeba71dbee984f42fe1270d"]}}`)
	d := Description{}
	err := d.New(descFile)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	newLayer := Layer{
		Size:   10,
		Sha256: "abc123",
	}
	d.AddLayer(&newLayer)
	// Last history entry should be the new layer
	if d.History[len(d.History)-1].Comment != "Created by Kubeless" {
		t.Errorf("Failed to include new layer: %v", d.History)
	}
	// Last rootfs.diff_id should be the new layer
	if d.Rootfs.DiffIds[len(d.Rootfs.DiffIds)-1] == "abc123" {
		t.Error("Failed to include new layer")
	}
}

func TestDescriptionToLayer(t *testing.T) {
	emptyDesc := Description{}
	res, err := emptyDesc.ToLayer()
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	expectedSize := int64(721)
	expectedSha := "17263670d4f12e26a270c7ec0a443c3ba8354da1d42f43f8e421634c5965bb6b"
	if res.Sha256 != expectedSha {
		t.Errorf("Unexpected sha256 %s", res.Sha256)
	}
	if res.Size != expectedSize {
		t.Errorf("Unexpected size %d", res.Size)
	}
}
