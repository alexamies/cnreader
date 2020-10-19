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
	"encoding/csv"
	"fmt"
	"io"

	"github.com/alexamies/cnreader/corpus"
)

// The library file listing the corpora
const LibraryFile = "data/corpus/library.csv"

type CorpusData struct {
	Title, ShortName, Status, FileName string
}

type Corpus struct {
	Title, Summary, DateUpdated string
	Collections []corpus.CollectionEntry
}

// A Library is a set of corpora loaded using a LibraryLoader and metadata
type Library struct {
	Title, Summary, DateUpdated, TargetStatus string
	Loader LibraryLoader
}
// A LibraryData is a struct to output library metadata to a HTML file
type LibraryData struct {
	Title, Summary, DateUpdated, TargetStatus string
	Corpora []CorpusData
}

// LibraryLoader loads teh corpora into the library
type LibraryLoader interface {

	// GetCorpusLoader gets the corpus loader
	GetCorpusLoader() corpus.CorpusLoader

	// LoadLibrary loads the corpora in the library
	LoadLibrary(r io.Reader) (*[]CorpusData, error)
}

// fileLibraryLoader loads the corpora from files
type fileLibraryLoader struct{
	FileName string
	Config corpus.CorpusConfig
	CLoader corpus.CorpusLoader
}

// GetCorpusLoader implements theLibraryLoader interface
func (loader fileLibraryLoader) GetCorpusLoader() corpus.CorpusLoader {
	return loader.CLoader
}

// LoadLibrary implements the interface
func (loader fileLibraryLoader) LoadLibrary(r io.Reader) (*[]CorpusData, error) {
	return loadLibrary(r)
}

// NewLibraryLoader creates a new LibraryLoader
func NewLibraryLoader(fname string, config corpus.CorpusConfig) LibraryLoader {
	return fileLibraryLoader{
		FileName: fname,
		Config: config,
		CLoader: corpus.NewCorpusLoader(config),
	}
}

// loadLibrary gets the list of source and destination files for HTML conversion
func loadLibrary(r io.Reader) (*[]CorpusData, error) {
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1
	reader.Comma = rune('\t')
	reader.Comment = rune('#')
	rawCSVdata, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("loadLibrary, error reading file: %v", err)
	}
	corpora := []CorpusData{}
	for i, row := range rawCSVdata {
		if len(row) < 4 {
			return nil, fmt.Errorf("library.loadLibrary: not enough rows in file %d, %d ", i,
				      len(row))
	  	}
		title := row[0]
		shortName := row[1]
		status := row[2]
		fileName := row[3]
		corpus := CorpusData{title, shortName, status, fileName}
		corpora = append(corpora, corpus)
	}
	return &corpora, nil
}
