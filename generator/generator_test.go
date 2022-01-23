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
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/alexamies/chinesenotes-go/config"
	"github.com/alexamies/chinesenotes-go/dicttypes"
	"github.com/alexamies/chinesenotes-go/tokenizer"
	"github.com/alexamies/cnreader/corpus"
)

func mockCorpusConfig() corpus.CorpusConfig {
	return corpus.CorpusConfig{
		CorpusDataDir: "data/corpus",
		CorpusDir:     "corpus",
		Excluded:      map[string]bool{},
		ProjectHome:   ".",
	}
}

func TestHyperlink(t *testing.T) {
	const vocabFormat = `<a title="%s | %s" class="%s" href="/words/%d.html">%s</a>`
	ws1 := dicttypes.WordSense{
		Pinyin:  "hǎi",
		English: "sea",
	}
	hw1 := dicttypes.Word{
		HeadwordId:  1,
		Simplified:  "海",
		Traditional: "\\N",
		Pinyin:      "hǎi",
		Senses:      []dicttypes.WordSense{ws1},
	}
	ws2 := dicttypes.WordSense{
		Pinyin:  "hǎi",
		English: "Simile of Medicinal Herbs",
		Notes:   "Parallel, Sanskrit equivalent: oṣadhīparivartaḥ",
	}
	hw2 := dicttypes.Word{
		HeadwordId:  2,
		Simplified:  "药草喻品",
		Traditional: "藥草喻品",
		Pinyin:      "yàocǎo yù pǐn",
		Senses:      []dicttypes.WordSense{ws2},
	}
	type test struct {
		name                    string
		word                    dicttypes.Word
		text, vocabFormat, want string
	}
	tests := []test{
		{
			name:        "Simple",
			word:        hw1,
			text:        "海",
			vocabFormat: vocabFormat,
			want:        `<a title="hǎi | sea" class="vocabulary" href="/words/1.html">海</a>`,
		},
		{
			name:        "Parallel",
			word:        hw2,
			text:        "藥草喻品",
			vocabFormat: vocabFormat,
			want:        `<a title="yàocǎo yù pǐn | Simile of Medicinal Herbs" class="vocabulary parallel" href="/words/2.html">藥草喻品</a>`,
		},
	}
	for _, tc := range tests {
		got := hyperlink(tc.word, tc.text, tc.vocabFormat)
		if got != tc.want {
			t.Errorf("%s got %s but want %s", tc.name, got, tc.want)
		}
	}
}

func TestSpan(t *testing.T) {
	s1 := "海"
	t1 := "\\N"
	ws1 := dicttypes.WordSense{
		Pinyin:  "hǎi",
		English: "sea",
	}
	hw1 := dicttypes.Word{
		HeadwordId:  1,
		Simplified:  s1,
		Traditional: t1,
		Pinyin:      "hǎi",
		Senses:      []dicttypes.WordSense{ws1},
	}
	s2 := "国"
	t2 := "國"
	ws2 := dicttypes.WordSense{
		Pinyin:  "guó",
		English: "country",
	}
	hw2 := dicttypes.Word{
		HeadwordId:  2,
		Simplified:  s2,
		Traditional: t2,
		Pinyin:      "guó",
		Senses:      []dicttypes.WordSense{ws2},
	}
	s3 := "菩萨"
	t3 := "菩薩"
	ws3 := dicttypes.WordSense{
		Pinyin:  "púsà",
		English: "bodhisattva",
		Notes:   "Sanskrit equivalent: bodhisattva",
	}
	hw3 := dicttypes.Word{
		HeadwordId:  3,
		Simplified:  s3,
		Traditional: t3,
		Pinyin:      "púsà",
		Senses:      []dicttypes.WordSense{ws3},
	}
	ws4 := dicttypes.WordSense{
		Notes: "FGDB entry 42",
	}
	hw4 := dicttypes.Word{
		HeadwordId:  3,
		Simplified:  s3,
		Traditional: t3,
		Pinyin:      "púsà",
		Senses:      []dicttypes.WordSense{ws3, ws4},
	}
	tests := []struct {
		name     string
		input    string
		hw       dicttypes.Word
		expected string
	}{
		{
			name:     "happy path",
			input:    "海",
			hw:       hw1,
			expected: `<span title="hǎi | sea" class="vocabulary" itemprop="HeadwordId" value="1">海</span>`,
		},
		{
			name:     "Tradition text",
			input:    "國",
			hw:       hw2,
			expected: `<span title="guó | country" class="vocabulary" itemprop="HeadwordId" value="2">國</span>`,
		},
		{
			name:     "Has Sanskrit",
			input:    "菩薩",
			hw:       hw3,
			expected: `<span title="púsà | bodhisattva" class="vocabulary sanskrit" itemprop="HeadwordId" value="3">菩薩</span>`,
		},
		{
			name:     "Has Sanskrit and is a FGDB entry",
			input:    "菩薩",
			hw:       hw4,
			expected: `<span title="púsà | 1. bodhisattva, 2. " class="vocabulary sanskrit fgdb" itemprop="HeadwordId" value="3">菩薩</span>`,
		},
	}
	for _, tc := range tests {
		highlighted := span(tc.hw, tc.input)
		if highlighted != tc.expected {
			t.Errorf("%s: expected %s, got %s", tc.name, tc.expected, highlighted)
		}
	}
}

func TestDecodeUsageExample(t *testing.T) {
	s1 := "海"
	t1 := "\\N"
	hw1 := dicttypes.Word{
		HeadwordId:  1,
		Simplified:  s1,
		Traditional: t1,
		Pinyin:      "hǎi",
		Senses:      []dicttypes.WordSense{},
	}
	s2 := "国"
	t2 := "國"
	hw2 := dicttypes.Word{
		HeadwordId:  2,
		Simplified:  s2,
		Traditional: t2,
		Pinyin:      "guó",
		Senses:      []dicttypes.WordSense{},
	}
	wdict := make(map[string]*dicttypes.Word)
	wdict[s1] = &hw1
	wdict[s2] = &hw2
	wdict[t2] = &hw2
	dictTokenizer := tokenizer.DictTokenizer{WDict: wdict}
	type testCase struct {
		name     string
		input    string
		hw       dicttypes.Word
		expected string
	}
	tests := []testCase{
		{
			name:     "happy path",
			input:    "海",
			hw:       hw1,
			expected: "<span class='usage-highlight'>海</span>",
		},
		{
			name:     "banana",
			input:    "banana",
			hw:       hw2,
			expected: "banana",
		},
		{
			name:     "Tradition text",
			input:    "國",
			hw:       hw2,
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
	wdict := make(map[string]*dicttypes.Word)
	const usageText = "繁"
	dictTokenizer := tokenizer.DictTokenizer{WDict: wdict}
	tokens := dictTokenizer.Tokenize(usageText)
	var buf bytes.Buffer
	vocab := make(map[string]int)
	noTemplates := make(map[string]*template.Template)
	outputConfig0 := HTMLOutPutConfig{
		Title:     "A title",
		Templates: noTemplates,
	}
	type test struct {
		name      string
		config    HTMLOutPutConfig
		expectErr bool
	}
	tests := []test{
		{
			name:      "No templates",
			config:    outputConfig0,
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
	wdict := make(map[string]*dicttypes.Word)
	tokenizer := tokenizer.DictTokenizer{WDict: wdict}
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
		name        string
		input       string
		expectError bool
		expectLen   int
	}
	tests := []test{
		{
			name:        "One character",
			input:       "繁",
			expectError: false,
			expectLen:   1,
		},
		{
			name:        "Four characters",
			input:       input1,
			expectError: false,
			expectLen:   6,
		},
		{
			name:        "Simplified characters",
			input:       input2,
			expectError: false,
			expectLen:   8,
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
	ch1 := corpus.CorpusEntry{
		RawFile:   "chapter1.txt",
		GlossFile: "chapter1.html",
		Title:     "Chapte 1",
		ColTitle:  "A Book",
	}
	colEntry := corpus.CollectionEntry{
		CollectionFile: "",
		GlossFile:      "",
		Title:          "A Book",
		Summary:        "A very good book",
		Intro:          "An intro",
		DateUpdated:    time.Now().Format("2006-01-02"),
		Corpus:         "",
		CorpusEntries:  []corpus.CorpusEntry{ch1},
		AnalysisFile:   "",
		Format:         "Prose",
		Date:           "100 BCE",
		Genre:          "Historic",
	}
	outputConfig := HTMLOutPutConfig{
		Title:            "My App",
		Templates:        NewTemplateMap(config.AppConfig{}),
		ContainsByDomain: "",
		Domain:           "",
		GoStaticDir:      "web",
		TemplateDir:      "",
		VocabFormat:      "",
		WebDir:           "web-staging",
	}
	type test struct {
		name          string
		entry         corpus.CollectionEntry
		wantToInclude string
	}
	tests := []test{
		{
			name:          "Empty",
			entry:         corpus.CollectionEntry{},
			wantToInclude: "",
		},
		{
			name:          "One entry",
			entry:         colEntry,
			wantToInclude: "web/chapter1.html",
		},
	}
	for _, tc := range tests {
		var buf bytes.Buffer
		err := WriteCollectionFile(&tc.entry, outputConfig, corpus.CorpusConfig{}, &buf)
		if err != nil {
			t.Fatalf("%s, unexpected error: %v", tc.name, err)
		}
		got := buf.String()
		if len(got) == 0 {
			t.Fatalf("%s, no data written", tc.name)
		}
		if !strings.Contains(got, tc.wantToInclude) {
			t.Errorf("%s, got %s\n but wantToInclude %s", tc.name, got, tc.wantToInclude)
		}
	}
	t.Log("generator.TestWriteDoc: End +++++++++++")
}

func TestWriteCollectionList(t *testing.T) {
	t.Log("generator.WriteCollectionList: Begin +++++++++++")
	type test struct {
		name    string
		entries []corpus.CollectionEntry
	}
	tests := []test{
		{
			name:    "Empty",
			entries: []corpus.CollectionEntry{},
		},
	}
	for _, tc := range tests {
		var buf bytes.Buffer
		outputConfig := HTMLOutPutConfig{}
		outputConfig.Templates = NewTemplateMap(config.AppConfig{})
		err := WriteCollectionList(tc.entries, "analysisFile",
			outputConfig, &buf)
		if err != nil {
			t.Fatalf("%s unexpected error: %v", tc.name, err)
		}
		if len(buf.String()) == 0 {
			t.Errorf("%s no data written", tc.name)
		}
	}
	t.Log("generator.WriteCollectionList: End +++++++++++")
}
