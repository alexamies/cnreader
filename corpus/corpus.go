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
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type CollectionEntry struct {
	CollectionFile, GlossFile, Title, Summary, Intro, DateUpdated, Corpus string
	CorpusEntries []CorpusEntry
	AnalysisFile, Format, Date, Genre string
}

const COLLECTIONS_FILE = "collections.csv"

// An entry in a collection
type CorpusEntry struct {
	RawFile, GlossFile, Title, ColTitle string
}

// CorpusConfig encapsulates parameters for corpus configuration
type CorpusConfig struct {
	CorpusDataDir string
	CorpusDir string
	Excluded map[string]bool
	ProjectHome string
}

type CorpusLoader interface {

	// Method to get a single entry in a collection
	// Param:
	//   fName: The file name of the collection
	// Returns
	//   A CollectionEntry encapsulating the collection or an error
	GetCollectionEntry(fName string) (CollectionEntry, error)

	// Load all entries in all collections in a corpus
    // Param:
	//  fName: the corpus file name listing the collections
	LoadAll(fName string) map[string]CorpusEntry

	// Method to load the entries in a collection
    // Param:
    //   fName: A file name containing the entries in the collection
    //   colTitle: The title of the collection
	LoadCollection(fName string, colTitle string) []CorpusEntry

	// Method to load the collections in a corpus
	// Parameter:
	//  fName: the corpus file name listing the collections
	LoadCorpus(fName string) []CollectionEntry

	// Method to read the contents of a corpus entry
	// Parameter:
	//  fName: the file name containing the text
	ReadText(fName string) string

}

// A FileLibraryLoader loads the corpora from files
type FileCorpusLoader struct{
	FileName string
	Config CorpusConfig
}

// Impements the CollectionLoader interface for FileCollectionLoader
func (loader FileCorpusLoader) GetCollectionEntry(fName string) (CollectionEntry, error) {
	return getCollectionEntry(loader.FileName, loader.Config)
}

// Implements the LoadAll method in the CorpusLoader interface
func (loader FileCorpusLoader) LoadAll(fName string) (map[string]CorpusEntry) {
	return loadAll(loader, fName)
}

// Implements the LoadCorpus method in the CorpusLoader interface
func (loader FileCorpusLoader) LoadCollection(fName, colTitle string) []CorpusEntry {
	return loadCorpusEntries(fName, colTitle, loader.Config)
}

// LoadCorpus implements the CorpusLoader interface
func (loader FileCorpusLoader) LoadCorpus(fName string) []CollectionEntry {
	return loadCorpusCollections(fName, loader.Config)
}

// Implements the LoadCorpus method in the CorpusLoader interface
func (loader FileCorpusLoader) ReadText(fName string) string {
	return readText(fName)
}

// Gets the entry the collection
// Parameter
// collectionFile: The name of the file describing the collection
func getCollectionEntry(collectionFile string, corpusConfig CorpusConfig) (CollectionEntry, error)  {
	log.Printf("corpus.GetCollectionEntry: collectionFile: '%s'.\n",
		collectionFile)
	collections := loadCorpusCollections(COLLECTIONS_FILE, corpusConfig)
	for _, entry := range collections {
		if strings.Compare(entry.CollectionFile, collectionFile) == 0 {
			return entry, nil
		}
	}
	return CollectionEntry{}, errors.New("could not find collection " +
		collectionFile)
}

// Method to get a a map of entries with keys being output (HTML) file names
// Param:
//   sourceMap: A map by source (plain) text file name
// Returns
//   map with keys being the output file names
func GetOutfileMap(sourceMap map[string]CorpusEntry) map[string]CorpusEntry {
	outMap := map[string]CorpusEntry{}
	for _, entry := range sourceMap {
		outMap[entry.GlossFile] = entry
	}
	return outMap
}

// Load all corpus entries and keep them in a hash map
func loadAll(loader CorpusLoader, fName string) (map[string]CorpusEntry) {
	corpusEntryMap := map[string]CorpusEntry{}
	collections := loader.LoadCorpus(fName)
	for _, collectionEntry := range collections {
		corpusEntries := loader.LoadCollection(collectionEntry.CollectionFile,
				collectionEntry.Title)
		for _, entry := range corpusEntries {
			corpusEntryMap[entry.RawFile] = entry
		}
	}
	return corpusEntryMap
}

// Gets the list of collections in the corpus
func loadCorpusCollections(cFile string, corpusConfig CorpusConfig) []CollectionEntry {
	log.Printf("corpus.loadCorpusCollections: cFile: '%s'.\n", cFile)
	collectionsFile := corpusConfig.CorpusDataDir + "/" + cFile
	file, err := os.Open(collectionsFile)
	if err != nil {
		log.Fatal("loadCorpusCollections: Error opening collection file.", err)
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
	collections := make([]CollectionEntry, 0)
	for i, row := range rawCSVdata {
		if len(row) < 9 {
			log.Fatal("loadCorpusCollections: not enough rows in file ", i,
					len(row), collectionsFile)
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
	return collections
}

// Get a list of files for a corpus
func loadCorpusEntries(collectionFile, colTitle string, corpusConfig CorpusConfig) []CorpusEntry {
	//log.Printf("corpus.loadCorpusEntries enter: '%s'.\n", collectionFile)
	cFile := corpusConfig.CorpusDataDir + "/" + collectionFile
	file, err := os.Open(cFile)
	if err != nil {
		log.Fatal("loadCorpusEntries collectionFile not found: ", err)
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
	corpusEntries := make([]CorpusEntry, 0)
	for _, row := range rawCSVdata {
		if len(row) != 3 {
			log.Fatal("corpus.loadCorpusEntries len(row) != 3 ", row)
		}
		corpusEntries = append(corpusEntries, CorpusEntry{row[0], row[1],
			row[2], colTitle})
	}
	return corpusEntries	
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
// introFile: The name of the file introducing the collection
func ReadIntroFile(introFile string, corpusConfig CorpusConfig) string {
	//log.Printf("ReadIntroFile: Reading introduction file.\n")
	infile, err := os.Open(corpusConfig.CorpusDataDir + "/" + introFile)
	if err != nil {
		log.Fatal(err)
	}
	reader := bufio.NewReader(infile)
	var buffer bytes.Buffer
	eof := false
	for !eof {
		var line string
		line, err = reader.ReadString('\n')
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
func readText(filename string) string {
	var text string
	if strings.HasSuffix(filename, ".html") {
		bs, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Printf("corpus.readText: could not read %s\n", filename)
			log.Fatal(err)
		}
		text = string(bs)
	} else { // plain text file, add line breaks
		infile, err := os.Open(filename)
		if err != nil {
			log.Fatal(err)
		}
		defer infile.Close()
		reader := bufio.NewReader(infile)
		var buffer bytes.Buffer
		eof := false
		for !eof {
			var line string
			line, err = reader.ReadString('\n')
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
		text = buffer.String()
	}
	//fmt.Printf("ReadText: read text %s\n", text)
	return text
}
