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

package corpus

import (
	"bytes"
	"fmt"
	"io"
	"testing"
	"time"
)

// Implements the CorpusLoader interface with trivial implementation
type EmptyCorpusLoader struct {Label string}

func (loader EmptyCorpusLoader) GetConfig() CorpusConfig {
	return CorpusConfig{}
}

func (loader EmptyCorpusLoader) GetCollectionEntry(fName string) (*CollectionEntry, error) {
	return &CollectionEntry{}, nil
}

func (loader EmptyCorpusLoader) LoadAll(r io.Reader) (*map[string]CorpusEntry, error) {
	return loadAll(loader, r)
}

func (loader EmptyCorpusLoader) LoadCollection(r io.Reader, colTitle string) (*[]CorpusEntry, error) {
	return nil, fmt.Errorf("not implemented")
}

func (loader EmptyCorpusLoader) LoadCorpus(r io.Reader) (*[]CollectionEntry, error) {
	return &[]CollectionEntry{}, nil
}

func (loader EmptyCorpusLoader) ReadText(r io.Reader) string {
	return ""
}

// Implements the CorpusLoader interface with mock data
type MockCorpusLoader struct {Label string}

func (loader MockCorpusLoader) GetCollectionEntry(fName string) (*CollectionEntry, error) {
	entry := CorpusEntry{
		RawFile: "corpus_doc.txt",
		GlossFile: "corpus_doc.html",
		Title: "corpus doc title",
		ColTitle: "A Corpus Collection",
	}
	dateUpdated := time.Now().Format("2006-01-02")
	c := CollectionEntry{
		CollectionFile: "a_collection_file.txt",
		GlossFile: "a_collection_file.html",
		Title: "A Corpus Collection",
		Summary: "A summary",
		Intro: "An introduction",
		DateUpdated: dateUpdated,
		Corpus: "A Corpus",
		CorpusEntries: []CorpusEntry{entry},
		AnalysisFile: "collection_analysis_file.html",
		Format: "prose",
		Date: "1984",
		Genre: "Science Fiction",
	}
	return &c, nil
}

func (loader MockCorpusLoader) GetConfig() CorpusConfig {
	return CorpusConfig{}
}

func (loader MockCorpusLoader) LoadAll(r io.Reader) (*map[string]CorpusEntry, error) {
	return loadAll(loader, r)
}

func (loader MockCorpusLoader) LoadCollection(r io.Reader, colTitle string) (*[]CorpusEntry, error) {
	entry := CorpusEntry{
		RawFile: "raw_file.txt",
		GlossFile: "gloss_file.html",
		Title: "Entry Title",
		ColTitle: "Collection Title",
	}
	return &[]CorpusEntry{entry}, nil
}

func (loader MockCorpusLoader) LoadCorpus(r io.Reader) (*[]CollectionEntry, error) {
	c := CollectionEntry{}
	return &[]CollectionEntry{c}, nil
}

func (loader MockCorpusLoader) ReadText(r io.Reader) string {
	return "你好 Hello!"
}

func mockCorpusConfig() CorpusConfig {
	return CorpusConfig{
		CorpusDataDir: "data/corpus",
		CorpusDir: "corpus",
		Excluded: map[string]bool{},
		ProjectHome: "..",
	}
}

func mockFileCorpusLoader() FileCorpusLoader {
	return FileCorpusLoader{
		FileName: "File",
		Config: mockCorpusConfig(),
	}
}

// Trivial test to load a collection file
func TestLoadAll(t *testing.T) {
	t.Log("corpus.TestLoadAll: Begin unit test")
	loader := EmptyCorpusLoader{"File"}
	simpleCol := `#Comment
example_collection.tsv	example_collection.html	A Book 一本書		example_collection000.txt		Literary Chinese	Prose	Classical
`
	tests := []struct {
		name string
		input string
		expectedLen int
	}{
		{
			name: "empty",
			input: "",
			expectedLen: 0,
		},
		{
			name: "simple collection",
			input: simpleCol,
			expectedLen: 1,
		},
	}
	for _, tc := range tests {
		var buf bytes.Buffer
		io.WriteString(&buf, tc.input)
		corpusEntryMap, err := loader.LoadAll(&buf)
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", tc.name, err)
		}
		if tc.expectedLen != len(*corpusEntryMap) {
			t.Errorf("%s: expectedLen %d, got: %d", tc.name, tc.expectedLen,
					len(*corpusEntryMap))
		}
	}
}

// Test reading of files for HTML conversion
func TestCollections(t *testing.T) {
	t.Log("corpus.TestCollections: Begin unit test")
	var buf bytes.Buffer
	collections, err := loadCorpusCollections(&buf, mockCorpusConfig())
	if err != nil {
		t.Fatalf("%s: unexpected error: %v", "TestCollections", err)
	}
	if len(*collections) == 0 {
		t.Error("No collections found")
	} else {
		genre := "Confucian"
		colList := *collections
		if colList[0].Genre != genre {
			t.Error("Expected genre ", genre, ", got ", colList[0].Genre)
		}
	}
	t.Log("corpus.TestCollections: End unit test")
}

// Test reading of corpus files
func TestLoadCollection(t *testing.T) {
	t.Log("corpus.TestLoadCollection: Begin unit test")
	emptyLoader := EmptyCorpusLoader{"Empty"}
	smallCollection := `# Source file, HTML file, title
example_collection/example_collection001.txt	example_collection/example_collection001.html	第一回 Chapter 1
`
	tests := []struct {
		name string
		input string
		expectedLen int
	}{
		{
			name: "empty",
			input: "",
			expectedLen: 0,
		},
		{
			name: "simple collection",
			input: smallCollection,
			expectedLen: 1,
		},
	}
	for _, tc := range tests {
		var buf bytes.Buffer
		io.WriteString(&buf, tc.input)
		corpusEntries, err := emptyLoader.LoadCollection(&buf, "")
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", tc.name, err)
		}
		if tc.expectedLen != len(*corpusEntries) {
			t.Errorf("%s: expectedLen %d, got: %d", tc.name, tc.expectedLen,
					len(*corpusEntries))
		}
	}
	t.Log("corpus.TestLoadCollection: End unit test")
}

// Test generating collection file
func TestReadIntroFile(t *testing.T) {
	ReadIntroFile("erya00.txt", mockCorpusConfig())
}
