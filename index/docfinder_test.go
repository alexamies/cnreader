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

package index

import (
	"testing"

	"github.com/alexamies/cnreader/corpus"
	"github.com/alexamies/chinesenotes-go/dicttypes"	
)

// IndexConfig encapsulates parameters for index configuration
func mockIndexConfig() IndexConfig {
	return IndexConfig{
		IndexDir: "index",
	}
}

func mockCorpusConfig() corpus.CorpusConfig {
	return corpus.CorpusConfig{
		CorpusDataDir: "data/corpus",
		CorpusDir: "corpus",
		Excluded: map[string]bool{},
		ProjectHome: "..",
	}
}

func mockDictionaryConfig() dicttypes.DictionaryConfig {
	return dicttypes.DictionaryConfig{
		AvoidSubDomains: map[string]bool{},
		DictionaryDir: "data",
	}
}

type mockValidator struct {}

func (mock mockValidator) Validate(pos, domain string) error {
	return nil
}

// Trivial test for document retrieval
func TestFindForKeyword0(t *testing.T) {
	BuildIndex(mockIndexConfig())
	documents := FindForKeyword("你")
	t.Logf("index.TestFindForKeyword0 %v", documents)
}

// Trivial test for loading index
func TestLoadKeywordIndex0(t *testing.T) {
	LoadKeywordIndex(mockIndexConfig())
}

// Trivial test for loading index
func TestFindDocsForKeyword0(t *testing.T) {
	BuildIndex(mockIndexConfig())
	s1 := "海"
	s2 := "\\N"
	hw := dicttypes.Word{
		HeadwordId:          1,
		Simplified:  s1,
		Traditional: s2,
		Pinyin:      "hǎi",
		Senses:  []dicttypes.WordSense{},
	}
	fileLoader := corpus.FileCorpusLoader{
		FileName: "File",
		Config: mockCorpusConfig(),
	}
	corpusEntryMap := fileLoader.LoadAll(corpus.CollectionsFile)
	outfileMap := corpus.GetOutfileMap(corpusEntryMap)
	documents := FindDocsForKeyword(hw, outfileMap)
	if len(documents) != 0 {
		t.Error("index.TestFindDocsForKeyword0: expectedd no documents")
	}
}

// Trivial test for loading index
func TestFindDocsForKeyword1(t *testing.T) {
	BuildIndex(mockIndexConfig())
	s1 := "铁"
	s2 := "鐵"
	hw := dicttypes.Word{
		HeadwordId:          1,
		Simplified:  s1,
		Traditional: s2,
		Pinyin:      "tiě",
		Senses:  []dicttypes.WordSense{},
	}
	fileLoader := corpus.FileCorpusLoader{
		FileName: "File",
		Config: mockCorpusConfig(),
	}
	corpusEntryMap := fileLoader.LoadAll(corpus.CollectionsFile)
	outfileMap := corpus.GetOutfileMap(corpusEntryMap)
	FindDocsForKeyword(hw, outfileMap)
}
