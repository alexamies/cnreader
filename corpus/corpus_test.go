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
	"testing"
)

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
func TestLoadAll0(t *testing.T) {
	t.Log("corpus.TestLoadAll: Begin unit test")
	loader := EmptyCorpusLoader{"File"}
	corpusEntryMap := loader.LoadAll(COLLECTIONS_FILE)
	if len(corpusEntryMap) != 0 {
		t.Error("corpus.TestLoadAll0: Non zero num. corpus entries found")
	}
}

// Easy test to load a collection file
func TestLoadAll1(t *testing.T) {
	t.Log("corpus.TestLoadAll1: Begin unit test")
	loader := MockCorpusLoader{"File"}
	corpusEntryMap := loader.LoadAll(COLLECTIONS_FILE)
	if len(corpusEntryMap) != 1 {
		t.Error("corpus.TestLoadAll1: No corpus entries found")
	}
	_, ok := corpusEntryMap["raw_file.txt"]
	if !ok {
		t.Logf("corpus.TestLoadAll1: corpusEntryMap %v", corpusEntryMap)
		t.Error("corpus.TestLoadAll1: Corpus entry not found")
	}
}

func TestLoadAll2(t *testing.T) {
	t.Log("corpus.TestLoadAll2: Begin unit test")
	fileLoader := mockFileCorpusLoader()
	corpusEntryMap := fileLoader.LoadAll(COLLECTIONS_FILE)
	if len(corpusEntryMap) == 0 {
		t.Error("corpus.TestLoadAll: No corpus entries found")
	} else {
		for _, v := range corpusEntryMap {
			entry := corpusEntryMap[v.RawFile]
			t.Fatalf("corpus.TestLoadAll2: first entry: %v", entry)
		}
	}
	t.Log("corpus.TestLoadAll: End unit test")
}

// Test reading of files for HTML conversion
func TestCollections(t *testing.T) {
	t.Log("corpus.TestCollections: Begin unit test")
	collections := loadCorpusCollections(COLLECTIONS_FILE, mockCorpusConfig())
	if len(collections) == 0 {
		t.Error("No collections found")
	} else {
		genre := "Confucian"
		if collections[0].Genre != genre {
			t.Error("Expected genre ", genre, ", got ",
				collections[0].Genre)
		}
	}
	t.Log("corpus.TestCollections: End unit test")
}

// Test reading of corpus files
func TestLoadCollection0(t *testing.T) {
	t.Log("corpus.TestLoadCollection0: Begin unit test")
	emptyLoader := EmptyCorpusLoader{"Empty"}
	corpusEntries := emptyLoader.LoadCollection("literary_chinese_prose.csv", "")
	if len(corpusEntries) != 0 {
		t.Error("Non zero corpus entries found")
	}
	t.Log("corpus.TestLoadCollection0: End unit test")
}

// Test reading of corpus files
func TestLoadCollection1(t *testing.T) {
	t.Log("corpus.TestLoadCollection1: Begin unit test")
	mockLoader := MockCorpusLoader{"Mock"}
	corpusEntries := mockLoader.LoadCollection("literary_chinese_prose.csv", "")
	if len(corpusEntries) != 1 {
		t.Error("Num corpus entries found != 1")
	}
	t.Log("corpus.TestLoadCollection1: End unit test")
}

// Test reading of corpus files
func TestLoadCollection2(t *testing.T) {
	fileLoader := mockFileCorpusLoader()
	corpusEntries := fileLoader.LoadCollection("literary_chinese_prose.csv", "")
	if len(corpusEntries) == 0 {
		t.Error("No corpus entries found")
	}
	if corpusEntries[0].RawFile != "classical_chinese_text-raw.html" {
		t.Error("Expected entry classical_chinese_text-raw.html, got ",
			corpusEntries[0].RawFile)
	}
	if corpusEntries[0].GlossFile != "classical_chinese_text.html" {
		t.Error("Expected entry classical_chinese_text.html, got ",
			corpusEntries[0].GlossFile)
	}
}

// Test generating collection file
func TestReadIntroFile(t *testing.T) {
	ReadIntroFile("erya00.txt", mockCorpusConfig())
}
