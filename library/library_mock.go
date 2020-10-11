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
	"github.com/alexamies/cnreader/corpus"
)

// Mock loader that has no data
type EmptyLibraryLoader struct {Label string}

// Implements the method from the LibraryLoader interface
func (loader EmptyLibraryLoader) GetCorpusLoader() corpus.CorpusLoader {
	return corpus.EmptyCorpusLoader{loader.Label}
}

func (loader EmptyLibraryLoader) LoadLibrary() []CorpusData {
	return []CorpusData{}
}

// Mock loader that generates static data
type MockLibraryLoader struct {Label string}

// Implements the method from the LibraryLoader interface
func (loader MockLibraryLoader) GetCorpusLoader() corpus.CorpusLoader {
	return corpus.MockCorpusLoader{loader.Label}
}

func (loader MockLibraryLoader) LoadLibrary() []CorpusData {
	c := CorpusData{
		Title: "Title",
		ShortName: "ShortName",
		Status: "Status",
		FileName: "FileName",
	}
	return []CorpusData{c}
}
