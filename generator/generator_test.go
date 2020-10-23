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
	"text/template"

	"github.com/alexamies/chinesenotes-go/config"
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

// TestNewTemplateMap building the template map
func TestNewTemplateMap(t *testing.T) {
	templates := NewTemplateMap(config.AppConfig{})
	_, ok := templates["corpus-template.html"]
	if !ok {
		t.Error("template not found")
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
	noTemplates := make(map[string]*template.Template)
	outputConfig0 := HTMLOutPutConfig{
		Title: "A title",
		Templates: noTemplates,
	}
	type test struct {
		name string
		config HTMLOutPutConfig
		expectErr bool
  }
  tests := []test{
		{
			name: "No templates",
			config: outputConfig0,
			expectErr: true,
		},
	}
	for _, tc := range tests {
 		err := WriteCorpusDoc(tokens, vocab, &buf, "", "", "", "", "TXT",
				tc.config, corpusConfig, wdict)
		if tc.expectErr && err == nil {
			t.Fatalf("%s: expected error but got none", tc.name)
		}
		if tc.expectErr {
			continue
		}
		if !tc.expectErr && err != nil {
			t.Fatalf("%s: unexpected error: %v", tc.name, err)
		}
	}
	t.Log("TestWriteCorpusDoc: End +++++++++++")
}

func TestWriteDoc(t *testing.T) {
	t.Log("generator.TestWriteDoc: Begin +++++++++++")
	wdict := make(map[string]dicttypes.Word)
	tokenizer := tokenizer.DictTokenizer{wdict}
	input1 := `
  	A test document
    繁體中文	
	`
	input2 := `
 	  <p>A test document with simplified Chinese</p>
    <p>简体中文</p>
    <p>Word with multiple senses: 中</p>
	`
	type test struct {
		name string
		input string
		expectError bool
		expectLen int
  }
  tests := []test{
		{
			name: "One character",
			input: "繁",
			expectError: false,
			expectLen: 1,
		},
		{
			name: "Four characters",
			input: input1,
			expectError: false,
			expectLen: 31,
		},
		{
			name: "Simplified characters",
			input: input2,
			expectError: false,
			expectLen: 109,
		},
  }
  for _, tc := range tests {
		tokens := tokenizer.Tokenize(tc.input)
		if tc.expectLen != len(tokens) {
			t.Errorf("%s, expected len %d, got %d", tc.name, tc.expectLen, len(tokens))
		}
		var buf bytes.Buffer
		tmpl := template.Must(template.New("page-html").Parse(``))
		const vocabFormat = `<details><summary>%s</summary>%s %s</details>`
		err := WriteDoc(tokens, &buf, *tmpl, true, "", vocabFormat, MarkVocabSummary)
		if !tc.expectError && err != nil {
			t.Errorf("%s unexpected error: %v", tc.name, err)
		}
	}
	t.Log("generator.TestWriteDoc: End +++++++++++")
}

func TestWriteCollectionFile(t *testing.T) {
	t.Log("generator.TestWriteCollectionFile: Begin +++++++++++")
	type test struct {
		name string
		entry corpus.CollectionEntry
  }
  tests := []test{
		{
			name: "Empty",
			entry: corpus.CollectionEntry{},
		},
  }
  for _, tc := range tests {
		var buf bytes.Buffer
		entries := []corpus.CorpusEntry{}
		outputConfig := HTMLOutPutConfig{}
		outputConfig.Templates = NewTemplateMap(config.AppConfig{})
		err := WriteCollectionFile(tc.entry, "corpus_analysis.html",
			outputConfig, corpus.CorpusConfig{}, entries, "introText", &buf)
		if err != nil {
			t.Errorf("%s unexpected error: %v", tc.name, err)
		}
		if len(buf.String()) == 0 {
			t.Errorf("%s no data written", tc.name)
		}
	}
	t.Log("generator.TestWriteDoc: End +++++++++++")
}
