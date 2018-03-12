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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

func copyReader(src io.Reader, dst string) error {
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, src)
	if err != nil {
		return err
	}
	err = dstFile.Sync()
	if err != nil {
		return err
	}
	return nil
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	return copyReader(srcFile, dst)
}

func addLayerFile(imageDir, tarFile string) (*Layer, error) {
	layerFile, err := os.Open(tarFile)
	if err != nil {
		return nil, err
	}
	defer layerFile.Close()
	// Copy tar layer to the image dir
	layer := Layer{}
	err = layer.New(layerFile)
	if err != nil {
		return nil, err
	}
	layerPath := path.Join(imageDir, fmt.Sprintf("%s.tar", layer.Sha256))
	if err != nil {
		return nil, err
	}
	err = copyFile(tarFile, layerPath)
	if err != nil {
		return nil, err
	}
	return &layer, nil
}

func updateDescription(descriptionDir string, descriptionFile *os.File, newLayer *Layer) (*Layer, error) {
	d := Description{}
	err := d.New(descriptionFile)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse image description: %v", err)
	}
	d.AddLayer(newLayer)
	if err != nil {
		return nil, fmt.Errorf("Unable to update image description: %v", err)
	}
	dLayer, err := d.ToLayer(descriptionDir)
	if err != nil {
		return nil, fmt.Errorf("Unable to generate layer from description: %v", err)
	}
	dContent, err := d.Content()
	if err != nil {
		return nil, fmt.Errorf("Unable to get description content: %v", err)
	}
	dLayerFile := path.Join(descriptionDir, fmt.Sprintf("%s.tar", dLayer.Sha256))
	err = copyReader(bytes.NewReader(dContent), dLayerFile)
	if err != nil {
		return nil, fmt.Errorf("Unable to store new description: %v", err)
	}
	return dLayer, nil
}

// AddTarToLayer copies a tar file into a image directory and update its metadata
func AddTarToLayer(imageDir, tarFile string) error {
	tarLayer, err := addLayerFile(imageDir, tarFile)
	if err != nil {
		return fmt.Errorf("Failed to copy tar file: %v", err)
	}

	// Parse manifest
	manifestPath := path.Join(imageDir, "manifest.json")
	manifestFile, err := os.Open(manifestPath)
	if err != nil {
		return err
	}
	m := Manifest{}
	err = m.New(manifestFile)
	if err != nil {
		return fmt.Errorf("Failed to parse image manifest: %v", err)
	}

	// Update description
	descriptionPath := path.Join(imageDir, fmt.Sprintf("%s.tar", strings.Replace(m.Config.Digest, "sha256:", "", -1)))
	descriptionFile, err := os.Open(descriptionPath)
	if err != nil {
		return err
	}
	descriptionLayer, err := updateDescription(imageDir, descriptionFile, tarLayer)
	if err != nil {
		return err
	}

	// Update manifest
	m.UpdateConfig(descriptionLayer)
	m.AddLayer(tarLayer)
	mBytes, err := json.Marshal(m)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(manifestPath, mBytes, 0644)
	if err != nil {
		return err
	}

	return nil
}
