/*
Copyright 2017 Google Inc. All rights reserved.

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

package jsonnet

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

// An Importer imports data from a path.
// TODO(sbarzowski) caching of errors (may require breaking changes)
type Importer interface {
	// Import fetches data from a given path. It may be relative
	// to the file where we do the import. What "relative path"
	// means depends on the importer.
	//
	// It is required that:
	// a) for given (importedFrom, importedPath) the same
	//    (contents, foundAt) are returned on subsequent calls.
	// b) for given foundAt, the contents are always the same
	//
	// It is recommended that if there are multiple locations that
	// need to be probed (e.g. relative + multiple library paths)
	// then all results of all attempts will be cached separately,
	// both nonexistence and contents of existing ones.
	// FileImporter may serve as an example.
	Import(importedFrom, importedPath string) (contents Contents, foundAt string, err error)
}

// Contents is a representation of imported data. It is a simple
// string wrapper, which makes it easier to enforce the caching policy.
type Contents struct {
	data *string
}

func (c Contents) String() string {
	return *c.data
}

// MakeContents creates Contents from a string.
func MakeContents(s string) Contents {
	return Contents{
		data: &s,
	}
}

// ImportCache represents a cache of imported data.
//
// While the user-defined Importer implementations
// are required to cache file content, this cache
// is an additional layer of optimization that caches values
// (i.e. the result of executing the file content).
// It also verifies that the content pointer is the same for two foundAt values.
type ImportCache struct {
	foundAtVerification map[string]Contents
	codeCache           map[string]potentialValue
	importer            Importer
}

// MakeImportCache creates an ImportCache using an Importer.
func MakeImportCache(importer Importer) *ImportCache {
	return &ImportCache{
		importer:            importer,
		foundAtVerification: make(map[string]Contents),
		codeCache:           make(map[string]potentialValue),
	}
}

func (cache *ImportCache) importData(importedFrom, importedPath string) (contents Contents, foundAt string, err error) {
	contents, foundAt, err = cache.importer.Import(importedFrom, importedPath)
	if err != nil {
		return Contents{}, "", err
	}
	if cached, importedBefore := cache.foundAtVerification[foundAt]; importedBefore {
		if cached != contents {
			panic(fmt.Sprintf("importer problem: a different instance of Contents returned when importing %#v again", foundAt))
		}
	} else {
		cache.foundAtVerification[foundAt] = contents
	}
	return
}

// ImportString imports a string, caches it and then returns it.
func (cache *ImportCache) ImportString(importedFrom, importedPath string, e *evaluator) (*valueString, error) {
	data, _, err := cache.importData(importedFrom, importedPath)
	if err != nil {
		return nil, e.Error(err.Error())
	}
	return makeValueString(data.String()), nil
}

func codeToPV(e *evaluator, filename string, code string) potentialValue {
	node, err := snippetToAST(filename, code)
	if err != nil {
		// TODO(sbarzowski) we should wrap (static) error here
		// within a RuntimeError. Because whether we get this error or not
		// actually depends on what happens in Runtime (whether import gets
		// evaluated).
		// The same thinking applies to external variables.
		return makeErrorThunk(err)
	}
	return makeThunk(makeInitialEnv(filename, e.i.baseStd), node)
}

// ImportCode imports code from a path.
func (cache *ImportCache) ImportCode(importedFrom, importedPath string, e *evaluator) (value, error) {
	contents, foundAt, err := cache.importData(importedFrom, importedPath)
	if err != nil {
		return nil, e.Error(err.Error())
	}
	var pv potentialValue
	if cachedPV, isCached := cache.codeCache[foundAt]; !isCached {
		// File hasn't been parsed and analyzed before, update the cache record.
		pv = codeToPV(e, foundAt, contents.String())
		cache.codeCache[foundAt] = pv
	} else {
		pv = cachedPV
	}
	return e.evaluate(pv)
}

// Concrete importers
// -------------------------------------

// FileImporter imports data from the filesystem.
type FileImporter struct {
	JPaths  []string
	fsCache map[string]*fsCacheEntry
}

type fsCacheEntry struct {
	exists   bool
	contents Contents
}

func (importer *FileImporter) tryPath(dir, importedPath string) (found bool, contents Contents, foundHere string, err error) {
	if importer.fsCache == nil {
		importer.fsCache = make(map[string]*fsCacheEntry)
	}
	var absPath string
	if path.IsAbs(importedPath) {
		absPath = importedPath
	} else {
		absPath = path.Join(dir, importedPath)
	}
	var entry *fsCacheEntry
	if cacheEntry, isCached := importer.fsCache[absPath]; isCached {
		entry = cacheEntry
	} else {
		contentBytes, err := ioutil.ReadFile(absPath)
		if err != nil {
			if os.IsNotExist(err) {
				entry = &fsCacheEntry{
					exists: false,
				}
			} else {
				return false, Contents{}, "", err
			}
		} else {
			entry = &fsCacheEntry{
				exists:   true,
				contents: MakeContents(string(contentBytes)),
			}
		}
		importer.fsCache[absPath] = entry
	}
	return entry.exists, entry.contents, absPath, nil
}

// Import imports file from the filesystem.
func (importer *FileImporter) Import(importedFrom, importedPath string) (contents Contents, foundAt string, err error) {
	dir, _ := path.Split(importedFrom)
	found, content, foundHere, err := importer.tryPath(dir, importedPath)
	if err != nil {
		return Contents{}, "", err
	}

	for i := len(importer.JPaths) - 1; !found && i >= 0; i-- {
		found, content, foundHere, err = importer.tryPath(importer.JPaths[i], importedPath)
		if err != nil {
			return Contents{}, "", err
		}
	}

	if !found {
		return Contents{}, "", fmt.Errorf("couldn't open import %#v: no match locally or in the Jsonnet library paths", importedPath)
	}
	return content, foundHere, nil
}

// MemoryImporter "imports" data from an in-memory map.
type MemoryImporter struct {
	Data map[string]Contents
}

// Import fetches data from a map entry.
// All paths are treated as absolute keys.
func (importer *MemoryImporter) Import(importedFrom, importedPath string) (contents Contents, foundAt string, err error) {
	if content, ok := importer.Data[importedPath]; ok {
		return content, importedPath, nil
	}
	return Contents{}, "", fmt.Errorf("import not available %v", importedPath)
}
