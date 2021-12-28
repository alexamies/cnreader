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
	"bytes"
	"strings"
	"testing"

	"github.com/alexamies/cnreader/corpus"
	"github.com/alexamies/chinesenotes-go/bibnotes"	
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

// Trivial test for loading index
func TestFindDocsForKeyword(t *testing.T) {
	var wfFile bytes.Buffer
	var wfDocReader bytes.Buffer
	var indexWriter bytes.Buffer
	indexStore := IndexStore{&wfFile, &wfDocReader, &indexWriter}
	indexState, err := BuildIndex(mockIndexConfig(), indexStore)
	if err != nil {
		t.Fatalf("index.TestFindDocsForKeyword: error building index: %v", err)
	}
	s1 := "海"
	s2 := "\\N"
	hw := dicttypes.Word{
		HeadwordId:          1,
		Simplified:  s1,
		Traditional: s2,
		Pinyin:      "hǎi",
		Senses:  []dicttypes.WordSense{},
	}
	corpusLoader := mockCorpusLoader{}
	outfileMap, err := corpus.GetOutfileMap(corpusLoader)
	if err != nil {
		t.Fatalf("index.TestFindDocsForKeyword: error: %v", err)
	}
  ref2FileReader := strings.NewReader("")
  refNo2ParallelReader := strings.NewReader("")
  refNo2TransReader := strings.NewReader("")
  bibNotesClient, err := bibnotes.LoadBibNotes(ref2FileReader, refNo2ParallelReader, refNo2TransReader)
  if err != nil {
  	t.Fatalf("TestFindDocsForKeyword error loading bibnotes: %v", err)
  }
	documents := FindDocsForKeyword(hw, *outfileMap, *indexState, bibNotesClient)
	if len(documents) != 0 {
		t.Errorf("index.TestFindDocsForKeyword: expected no documents: %d", len(documents))
	}
}
