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
	"github.com/alexamies/cnreader/corpus"
	"log"
	"os"
)

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

// A LibraryLoader loads teh corpora into the library
type LibraryLoader interface {

	// Method to get the corpus loader
	GetCorpusLoader() corpus.CorpusLoader

	// Method to load the corpora in the library
	LoadLibrary() []CorpusData
}

// A FileLibraryLoader loads the corpora from files
type FileLibraryLoader struct{
	FileName string
	Config corpus.CorpusConfig
}

// Implements the method from the LibraryLoader interface
func (loader FileLibraryLoader) GetCorpusLoader() corpus.CorpusLoader {
	return corpus.FileCorpusLoader{loader.FileName, loader.Config}
}

// Implements the method from the LibraryLoader interface
func (loader FileLibraryLoader) LoadLibrary() []CorpusData {
	return loadLibrary(loader.FileName)
}

// The library file listing the corpora
const LibraryFile = "data/corpus/library.csv"

// Gets the list of source and destination files for HTML conversion
func loadLibrary(fname string) []CorpusData {
	file, err := os.Open(fname)
	if err != nil {
		log.Fatal("library.loadLibrary: Error opening library file.", err)
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	reader.Comma = rune('\t')
	reader.Comment = rune('#')
	rawCSVdata, err := reader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}
	corpora := []CorpusData{}
	for i, row := range rawCSVdata {
		if len(row) < 4 {
			log.Fatal("library.loadLibrary: not enough rows in file ", i,
				      len(row), fname)
	  	}
		title := row[0]
		shortName := row[1]
		status := row[2]
		fileName := row[3]
		corpus := CorpusData{title, shortName, status, fileName}
		corpora = append(corpora, corpus)
	}
	return corpora
}
