// Test document retrieval
package index

import (
	"fmt"
	"github.com/alexamies/cnreader/corpus"
	"github.com/alexamies/cnreader/dictionary"
	"testing"
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

func mockDictionaryConfig() dictionary.DictionaryConfig {
	return dictionary.DictionaryConfig{
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
	fmt.Println("index.TestFindForKeyword0 ", documents)
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
	hw := dictionary.HeadwordDef{
		Id:          1,
		Simplified:  &s1,
		Traditional: &s2,
		Pinyin:      []string{"hǎi"},
		WordSenses:  &[]dictionary.WordSenseEntry{},
	}
	fileLoader := corpus.FileCorpusLoader{
		FileName: "File",
		Config: mockCorpusConfig(),
	}
	corpusEntryMap := fileLoader.LoadAll(corpus.COLLECTIONS_FILE)
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
	hw := dictionary.HeadwordDef{
		Id:          1,
		Simplified:  &s1,
		Traditional: &s2,
		Pinyin:      []string{"tiě"},
		WordSenses:  &[]dictionary.WordSenseEntry{},
	}
	fileLoader := corpus.FileCorpusLoader{
		FileName: "File",
		Config: mockCorpusConfig(),
	}
	corpusEntryMap := fileLoader.LoadAll(corpus.COLLECTIONS_FILE)
	outfileMap := corpus.GetOutfileMap(corpusEntryMap)
	FindDocsForKeyword(hw, outfileMap)
}
