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

package generator

import (
	"bytes"
	"testing"

	"github.com/alexamies/chinesenotes-go/dicttypes"	
	"github.com/alexamies/chinesenotes-go/tokenizer"
	"github.com/alexamies/cnreader/corpus"
)

func mockCorpusConfig() corpus.CorpusConfig {
	return corpus.CorpusConfig{
		CorpusDataDir: "data/corpus",
		CorpusDir: "corpus",
		Excluded: map[string]bool{},
		ProjectHome: ".",
	}
}

func TestDecodeUsageExample(t *testing.T) {
	s1 := "海"
	t1 := "\\N"
	hw1 := dicttypes.Word{
		HeadwordId:  	1,
		Simplified:  	s1,
		Traditional: 	t1,
		Pinyin:      	"hǎi",
		Senses:  			[]dicttypes.WordSense{},
	}
	s2 := "国"
	t2 := "國"
	hw2 := dicttypes.Word{
		HeadwordId:  	2,
		Simplified:  	s2,
		Traditional: 	t2,
		Pinyin:      	"guó",
		Senses:  			[]dicttypes.WordSense{},
	}
	wdict := make(map[string]dicttypes.Word)
	wdict[s1] = hw1
	wdict[s2] = hw2
	wdict[t2] = hw2
	dictTokenizer := tokenizer.DictTokenizer{wdict}
	type testCase struct {
		name string
		input string
		hw dicttypes.Word
		expected string
  }
  tests := []testCase{
		{
			name: "happy path",
			input: "海",
			hw: hw1,
			expected: "<span class='usage-highlight'>海</span>",
		},
		{
			name: "banana",
			input: "banana",
			hw: hw2,
			expected: "banana",
		},
		{
			name: "Tradition text",
			input: "國",
			hw: hw2,
			expected: "<span class='usage-highlight'>國</span>",
		},
	}
	outputConfig := HTMLOutPutConfig{}
  for _, tc := range tests {
		highlighted := DecodeUsageExample(tc.input, tc.hw, dictTokenizer,
				outputConfig, wdict)
		if highlighted != tc.expected {
			t.Errorf("%s: expected %s, got %s", tc.name, tc.expected, highlighted)
		}
	}
}

// TestWriteCorpusDoc tests the writeCorpusDoc function
func TestWriteCorpusDoc(t *testing.T) {
	t.Log("TestWriteCorpusDoc: Begin +++++++++++")
	corpusConfig := mockCorpusConfig()
	wdict := make(map[string]dicttypes.Word)
	const usageText = "繁"
	dictTokenizer := tokenizer.DictTokenizer{wdict}
	tokens := dictTokenizer.Tokenize(usageText)
	var buf bytes.Buffer
	vocab := make(map[string]int)
	outputConfig := HTMLOutPutConfig{}
	err := WriteCorpusDoc(tokens, vocab, &buf, "", "", "", "", "TXT",
			outputConfig, corpusConfig, wdict)
	if err != nil {
		t.Fatalf("TestWriteCorpusDoc error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("TestWriteCorpusDoc buffer is empty")
	}
	t.Log("TestWriteCorpusDoc: End +++++++++++")
}