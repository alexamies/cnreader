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
	"io"
	"testing"
	"time"
)

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
	return loadAll(loader)
}

func (loader MockCorpusLoader) LoadCollection(fName, colTitle string) (*[]CorpusEntry, error) {
	entry := CorpusEntry{
		RawFile: "raw_file.txt",
		GlossFile: "gloss_file.html",
		Title: "Entry Title",
		ColTitle: "Collection Title",
	}
	return &[]CorpusEntry{entry}, nil
}

func (loader MockCorpusLoader) LoadCollections() (*[]CollectionEntry, error) {
	entries := []CollectionEntry{}
	return &entries, nil
}

func (loader MockCorpusLoader) LoadCorpus(r io.Reader) (*[]CollectionEntry, error) {
	c := CollectionEntry{}
	return &[]CollectionEntry{c}, nil
}

func (loader MockCorpusLoader) ReadText(src string) (string, error) {
	return "你好 Hello!", nil
}

func mockCorpusConfig() CorpusConfig {
	return CorpusConfig{
		CorpusDataDir: "data/corpus",
		CorpusDir: "corpus",
		Excluded: map[string]bool{},
		ProjectHome: "..",
	}
}

// TestLoadAll tests loadAll
func TestLoadAll(t *testing.T) {
	mockCorpLoader := MockCorpusLoader{}
	tests := []struct {
		name string
	}{
		{
			name: "empty",
		},
	}
	for _, tc := range tests {
		_, err := loadAll(mockCorpLoader)
		if err != nil {
			t.Errorf("TestLoadAll, unexpected error, %s: %v", tc.name, err)
		}
	}
}

// LoadCorpus test to load a collection file
func TestLoadCorpus(t *testing.T) {
	t.Log("corpus.LoadCorpus: Begin unit test")
	loader := NewFileCorpusLoader(CorpusConfig{})
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
		corpusEntryMap, err := loader.LoadCorpus(&buf)
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
func TestLoadCorpusEntries(t *testing.T) {
	t.Log("corpus.TestCollections: Begin unit test")
	var buf bytes.Buffer
	collections, err := loadCorpusCollections(&buf)
	if err != nil {
		t.Fatalf("%s: unexpected error: %v", "TestCollections", err)
	}
	if len(*collections) != 0 {
		t.Errorf("More than zero collections found: %d", len(*collections))
	}
	t.Log("corpus.TestCollections: End unit test")
}

// Test reading of corpus files
func TestLoadCollection(t *testing.T) {
	t.Log("corpus.TestLoadCollection: Begin unit test")
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
		corpusEntries, err := loadCorpusEntries(&buf, "", "")
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
	expect := "令甲卒皆伏，使老弱女子乘城。"
	var buf bytes.Buffer
	io.WriteString(&buf, expect)
	got := ReadIntroFile(&buf)
	if (expect != got) {
		t.Errorf("Expected %s, got %s", expect, got)
	}
}
