// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package library

import (
	"fmt"
	"io"
	"testing"

	"github.com/alexamies/cnreader/corpus"
)

// Implements the CorpusLoader interface with trivial implementation
type EmptyCorpusLoader struct {Label string}

func (loader EmptyCorpusLoader) GetConfig() corpus.CorpusConfig {
	return corpus.CorpusConfig{}
}

func (loader EmptyCorpusLoader) GetCollectionEntry(fName string) (*corpus.CollectionEntry, error) {
	return &corpus.CollectionEntry{}, nil
}

func (loader EmptyCorpusLoader) LoadAll(r io.Reader) (*map[string]corpus.CorpusEntry, error) {
	entries := make(map[string]corpus.CorpusEntry)
	return &entries, nil
}

func (loader EmptyCorpusLoader) LoadCollection(r io.Reader, colTitle string) (*[]corpus.CorpusEntry, error) {
	return nil, fmt.Errorf("not implemented")
}

func (loader EmptyCorpusLoader) LoadCorpus(r io.Reader) (*[]corpus.CollectionEntry, error) {
	return &[]corpus.CollectionEntry{}, nil
}

func (loader EmptyCorpusLoader) ReadText(r io.Reader) string {
	return ""
}

// Mock loader that has no data
type EmptyLibraryLoader struct {Label string}

// Implements the method from the LibraryLoader interface
func (loader EmptyLibraryLoader) GetCorpusLoader() corpus.CorpusLoader {
	return EmptyCorpusLoader{loader.Label}
}

func (loader EmptyLibraryLoader) LoadLibrary() []CorpusData {
	return []CorpusData{}
}

func TestLoadLibrary(t *testing.T) {
	emptyLibLoader := EmptyLibraryLoader{"Empty"}
	emptyLibLoader.LoadLibrary()
}
