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

//Package for scanning the corpus collections
package corpus

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/gomarkdown/markdown"
)

const collectionsFile = "collections.csv"

type CollectionEntry struct {
	CollectionFile, GlossFile, Title, Summary, Intro, DateUpdated, Corpus string
	CorpusEntries []CorpusEntry
	AnalysisFile, Format, Date, Genre string
}

// An entry in a collection
type CorpusEntry struct {
	RawFile, GlossFile, Title, ColTitle, ColFile string
}

// CorpusConfig encapsulates parameters for corpus configuration
type CorpusConfig struct {
	CorpusDataDir string
	CorpusDir string
	Excluded map[string]bool
	ProjectHome string
	readCollections  func() (*[]CollectionEntry, error)
	readCollection  func(fName, colTitle string) (*[]CorpusEntry, error)
	readText func(string) (string, error)
}

// Interface for loading corpus with hierarchical collections of documents
type CorpusLoader interface {

	// Method to get the corpus configuration
	// Parameter:
	//  r: to reader the text
	GetConfig() CorpusConfig

	// Method to get a single entry in a collection
	// Param:
	//   fName: The file name of the collection
	// Returns
	//   A CollectionEntry encapsulating the collection or an error
	GetCollectionEntry(fName string) (*CollectionEntry, error)

	// Method to load the entries in a collection
    // Param:
    //   fName: A file name containing the entries in the collection
    //   colTitle: The title of the collection
	LoadCollection(fName, colTitle string) (*[]CorpusEntry, error)

	// Method to load the collections in a corpus from the default file
	// Parameter:
	//  r: to read the listing of the collections
	LoadCollections() (*[]CollectionEntry, error)

	// Method to load the collections in a corpus
	// Parameter:
	//  r: to read the listing of the collections
	LoadCorpus(r io.Reader) (*[]CollectionEntry, error)

	// Method to read the contents of a corpus entry
	// Parameter:
	//  r: to reader the text
	ReadText(srcFile string) (string, error)

}

// A FileLibraryLoader loads the corpora from files
type fileCorpusLoader struct{
	Config CorpusConfig
}

// Creates a new CorpusConfig strct
func NewFileCorpusConfig(corpusDataDir, corpusDir string,
		excluded map[string]bool, projectHome string) CorpusConfig {
	collectionsFile := corpusDataDir + "/" + collectionsFile

	readCollections := func() (*[]CollectionEntry, error) {
		f, err := os.Open(collectionsFile)
		if err != nil {
			return nil, fmt.Errorf("readCollections: Error opening file: %v", err)
		}
		defer f.Close()
		return loadCorpusCollections(f)
	}

	readCollection := func(fName, colTitle string) (*[]CorpusEntry, error) {
		collectionsFile := corpusDataDir + "/" + fName
		r, err := os.Open(collectionsFile)
		if err != nil {
			return nil, fmt.Errorf("readCollection: Error opening file %s: %v",
					collectionsFile, err)
		}
		defer r.Close()
		return loadCorpusEntries(r, colTitle, collectionsFile)
	}

	readTextImpl := func(srcFile string) (string, error) {
		src := corpusDir + "/" + srcFile
		reader, err := os.Open(src)
		if err != nil {
			return "", fmt.Errorf("readTextImpl: Error opening doc file %s: %v",
				src, err)
		}
		defer reader.Close()
		if strings.HasSuffix(srcFile, ".md") {
			md := ReadText(reader)
			return convertMarkdown(md), nil
		}
		return ReadText(reader), nil
	}

	return CorpusConfig{
		CorpusDataDir: corpusDataDir,
		CorpusDir: corpusDir,
		Excluded: excluded,
		ProjectHome: projectHome,
		readCollections: readCollections,
		readCollection: readCollection,
		readText: readTextImpl,
	}
}

// Impements the CollectionLoader interface for FileCollectionLoader
func (loader fileCorpusLoader) GetConfig() CorpusConfig {
	return loader.Config
}

// Impements the CollectionLoader interface for FileCollectionLoader
func (loader fileCorpusLoader) GetCollectionEntry(fName string) (*CollectionEntry, error) {
	return getCollectionEntry(fName, loader.Config)
}

// Implements the LoadCollection method in the CorpusLoader interface
func (loader fileCorpusLoader) LoadCollection(fName, colTitle string) (*[]CorpusEntry, error) {
	return loader.Config.readCollection(fName, colTitle)
}

// LoadCorpus implements the CorpusLoader interface
func (loader fileCorpusLoader) LoadCollections() (*[]CollectionEntry, error) {
	return loader.Config.readCollections()
}

// LoadCorpus implements the CorpusLoader interface
func (loader fileCorpusLoader) LoadCorpus(r io.Reader) (*[]CollectionEntry, error) {
	return loadCorpusCollections(r)
}

// Implements the LoadCorpus method in the CorpusLoader interface
func (loader fileCorpusLoader) ReadText(srcFile string) (string, error) {
	return loader.Config.readText(srcFile)
}

// CorpusLoader gets the default kind of CorpusLoader
func NewFileCorpusLoader(corpusConfig CorpusConfig) CorpusLoader  {
	return fileCorpusLoader{corpusConfig}
}

// Gets the entry the collection
// Parameter
// collectionFile: The name of the file describing the collection
func getCollectionEntry(collectionFile string, corpusConfig CorpusConfig) (*CollectionEntry, error)  {
	log.Printf("corpus.getCollectionEntry: collectionFile: '%s'.\n",
		collectionFile)
	cFile := corpusConfig.CorpusDataDir + "/" + collectionFile
	file, err := os.Open(cFile)
	if err != nil {
		log.Fatalf("getCollectionEntry: Error opening collection file: %v", err)
	}
	defer file.Close()
	collections, err := loadCorpusCollections(file)
	if err != nil {
		return nil, fmt.Errorf("getCollectionEntry count load collections: %v", err)
	}
	for _, entry := range *collections {
		if strings.Compare(entry.CollectionFile, collectionFile) == 0 {
			return &entry, nil
		}
	}
	return nil, fmt.Errorf("could not find collection: %v", collectionFile)
}

// Method to get a a map of entries with keys being output (HTML) file names
// Param:
//   sourceMap: A map by source (plain) text file name
// Returns
//   map with keys being the output file names
func GetOutfileMap(loader CorpusLoader) (*map[string]CorpusEntry, error) {
	sourceMap, err := loadAll(loader)
	if err != nil {
		return nil, fmt.Errorf("Unable to load corpus: %v", err)
	}
	outMap := map[string]CorpusEntry{}
	for _, entry := range *sourceMap {
		outMap[entry.GlossFile] = entry
	}
	return &outMap, nil
}

// Load all corpus entries and keep them in a hash map
func loadAll(loader CorpusLoader) (*map[string]CorpusEntry, error) {
	corpusEntryMap := map[string]CorpusEntry{}
	if loader == nil {
		log.Print("loadAll loader is nil")
		m := make(map[string]CorpusEntry)
		return &m, nil
	}
	collections, err := loader.LoadCollections()
	if err != nil {
		return nil, fmt.Errorf("loadAll could not load corpus: %v", err)
	}
	for _, collectionEntry := range *collections {
		colEntryFName := collectionEntry.CollectionFile
		corpusEntries, err := loader.LoadCollection(colEntryFName, collectionEntry.Title)
		if err != nil {
			return nil, fmt.Errorf("loadAll could not load collection %s: %v",
				collectionEntry.Title, err)
		}
		for _, entry := range *corpusEntries {
			corpusEntryMap[entry.RawFile] = entry
		}
	}
	return &corpusEntryMap, nil
}

// Gets the list of collections in the corpus
func loadCorpusCollections(r io.Reader) (*[]CollectionEntry, error) {
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1
	reader.Comma = rune('\t')
	reader.Comment = rune('#')
	rawCSVdata, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("loadCorpusCollections, could not read collections: %v", err)
	}
	collections := make([]CollectionEntry, 0)
	log.Printf("loadCorpusCollections, reading collections")
	for i, row := range rawCSVdata {
		//log.Printf("loadCorpusCollections, i = %d, len(row) = %d", i, len(row))
		if len(row) < 9 {
			return nil, fmt.Errorf("loadCorpusCollections: not enough fields in file line %d: %d",
					i, len(row))
	  }
		collectionFile := row[0]
		title := ""
		if row[2] != "\\N" {
			title = row[2]
		}
		summary := ""
		if row[3] != "\\N" {
			summary = row[3]
		}
		introFile := ""
		if row[4] != "\\N" {
			introFile = row[4]
		}
		corpus := ""
		if row[5] != "\\N" {
			corpus = row[5]
		}
		format := ""
		if row[6] != "\\N" {
			format = row[6]
		}
		date := ""
		if row[7] != "\\N" {
			date = row[7]
		}
		genre := ""
		if len(row) > 8 && row[8] != "\\N" {
			genre = row[8]
		}
		corpusEntries := make([]CorpusEntry, 0)
		// log.Printf("corpus.Collections: Read collection %s in corpus %s\n",
		//	collectionFile, corpus)
		collections = append(collections, CollectionEntry{collectionFile,
			row[1], title, summary, introFile, "", corpus, corpusEntries, "",
			format, date, genre})
	}
	return &collections, nil
}

// Get a list of files for a corpus
func loadCorpusEntries(r io.Reader, colTitle, colFile string) (*[]CorpusEntry, error) {
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1
	reader.Comma = rune('\t')
	reader.Comment = rune('#')
	rawCSVdata, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("loadCorpusEntries, unable to read corpus document: %v", err)
	}
	corpusEntries := make([]CorpusEntry, 0)
	for _, row := range rawCSVdata {
		if len(row) != 3 {
			return nil, fmt.Errorf("corpus.loadCorpusEntries len(row) != 3: %s", row)
		}
		entry := CorpusEntry{
			RawFile: row[0],
			GlossFile: row[1],
			Title: row[2],
			ColTitle: colTitle,
			ColFile: colFile,
		}
		corpusEntries = append(corpusEntries, entry)
	}
	return &corpusEntries, nil
}

// Constructor for an empty CollectionEntry
func NewCorpusEntry() *CorpusEntry {
	return &CorpusEntry{
		RawFile: "",
		GlossFile: "",
		Title: "",
	}
}

// Reads a text file introducing the collection. The file should be a plain
// text file. HTML breaks will be added for line breaks.
// Parameter
// r: with text introducing the collection
func ReadIntroFile(r io.Reader,) string {
	reader := bufio.NewReader(r)
	var buffer bytes.Buffer
	eof := false
	for !eof {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			err = nil
			eof = true
		} else if err != nil {
			break
		}
		if _, err = buffer.WriteString(line); err != nil {
			break
		}
	}
	return buffer.String()
}

// Reads a Chinese text file
func ReadText(r io.Reader) string {
	reader := bufio.NewReader(r)
	var buffer bytes.Buffer
	eof := false
	for !eof {
		var line string
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			err = nil
			eof = true
		} else if err != nil {
			break
		}
		if _, err = buffer.WriteString(line); err != nil {
			break
		}
	}
	return buffer.String()
}

func convertMarkdown(md string) string {
	b := markdown.ToHTML([]byte(md), nil, nil)
	return string(b)
}