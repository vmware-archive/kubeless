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
	"fmt"
	"io/ioutil"
	"os"
)

// Layer represent the size and checksum of a image layer
type Layer struct {
	Size   int64
	Sha256 string
}

// New returns a Layer based on its file
func (f *Layer) New(layerFile *os.File) error {
	// Calculate sha256
	fContent, err := ioutil.ReadAll(layerFile)
	if err != nil {
		return err
	}
	f.Sha256 = fmt.Sprintf("%x", sha256.Sum256(fContent))

	// Calculate size
	fstat, err := layerFile.Stat()
	if err != nil {
		return err
	}
	f.Size = fstat.Size()
	return nil
}
